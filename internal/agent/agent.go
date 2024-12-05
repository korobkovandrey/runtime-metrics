package agent

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/service"
)

type cfgTimeouts struct {
	pollInterval      time.Duration
	reportInterval    time.Duration
	httpTimeout       time.Duration
	sendExpireTimeout time.Duration
	dataExpireTimeout time.Duration
}

type sentMessage struct {
	sent  time.Time
	t     string
	name  string
	value string
}

type reportMessage struct {
	expire time.Time
	t      string
	name   string
	value  string
}

type Agent struct {
	sentChan      chan sentMessage
	reportChan    chan reportMessage
	gaugeSource   *service.Source
	collectExpire time.Time
	config        config.Config
	cfgTimeouts   cfgTimeouts
}

func (a *Agent) makeURL(r reportMessage) (string, error) {
	switch r.t {
	case service.GaugeType:
		return a.config.UpdateURL + service.GaugeType + "/" + r.name + "/" + r.value, nil
	case service.CounterType:
		return a.config.UpdateURL + service.CounterType + "/" + r.name + "/" + r.value, nil
	default:
		return "", fmt.Errorf("%s: %w", r.t, errors.New("type is not valid"))
	}
}

func (a *Agent) dataWorker() {
	defer close(a.reportChan)
	var (
		err    error
		expire time.Time
	)
	for {
		if len(a.sentChan) > 0 {
			var sent []service.DataSent
			for m := range a.sentChan {
				sent = append(sent, service.DataSent{
					T:     m.t,
					Name:  m.name,
					Value: m.value,
					Sent:  m.sent,
				})
				if len(a.sentChan) == 0 {
					break
				}
			}
			if len(sent) > 0 {
				a.gaugeSource.SetDataSent(sent)
			}
		}

		// получение данных раз pollInterval
		if a.collectExpire.Before(time.Now()) {
			a.collectExpire = time.Now().Add(a.cfgTimeouts.pollInterval)
			err = a.gaugeSource.Collect()
			if err != nil {
				log.Println(fmt.Errorf("dataWorker: %w", err))
			}
		}

		// метрики, которые не отправлялись дольше reportInterval
		dataForSend := a.gaugeSource.GetDataForSend(a.cfgTimeouts.dataExpireTimeout, a.cfgTimeouts.reportInterval)

		if len(dataForSend) > 0 {
			expire = time.Now().Add(a.cfgTimeouts.sendExpireTimeout)
			for _, v := range dataForSend {
				a.reportChan <- reportMessage{
					expire: expire,
					t:      v.T,
					name:   v.Name,
					value:  v.Value,
				}
			}
		}
		// раз в секунду
		time.Sleep(time.Second)
	}
}

func (a *Agent) reportWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	var now time.Time

	client := &http.Client{
		Timeout: a.cfgTimeouts.httpTimeout,
	}
	for c := range a.reportChan {
		if c.expire.Before(time.Now()) {
			continue
		}
		url, err := a.makeURL(c)
		if err != nil {
			log.Println(err)
			continue
		}

		now = time.Now()
		// @todo сделать в func sendRequest(client *http.Client) (ok bool, err error)
		response, err := client.Post(url, "text/plain", http.NoBody)
		if err != nil {
			log.Println(err)
		}
		if response == nil {
			continue
		}
		body, err := io.ReadAll(response.Body)
		err = errors.Join(err, response.Body.Close())
		if err != nil {
			log.Println(err)
		}
		if response.StatusCode == http.StatusOK {
			a.sentChan <- sentMessage{
				sent:  now,
				t:     c.t,
				name:  c.name,
				value: c.value,
			}
		} else {
			log.Println(url, response.StatusCode, strings.TrimSuffix(string(body), "\n"))
		}
	}
}

func New(cfg *config.Config) *Agent {
	gaugeSource := service.NewGaugeSource()
	gaugeSourceLen := gaugeSource.Len()

	a := &Agent{
		sentChan:    make(chan sentMessage, gaugeSourceLen),
		reportChan:  make(chan reportMessage, gaugeSourceLen),
		gaugeSource: gaugeSource,
		config:      *cfg,
		cfgTimeouts: cfgTimeouts{
			pollInterval:   time.Duration(cfg.PollInterval) * time.Second,
			reportInterval: time.Duration(cfg.ReportInterval) * time.Second,
		},
	}

	return a
}

func (a *Agent) Run() error {
	if a.cfgTimeouts.reportInterval <= a.cfgTimeouts.pollInterval {
		return fmt.Errorf("reportInterval (%v) must be greater than pollInterval (%v)",
			a.cfgTimeouts.reportInterval, a.cfgTimeouts.pollInterval)
	}
	if a.config.TimeoutCoefficient > 1 {
		return fmt.Errorf("TimeoutCoefficient (%f) must be less than 1 ", a.config.TimeoutCoefficient)
	}

	// максимальное время на запрос, чтобы успеть уложить отправку всех метрик в reportInterval
	to := float64(a.cfgTimeouts.reportInterval) / float64(a.gaugeSource.Len()) * float64(a.config.ReportWorkersCount)

	// таймаут для http клиента делаем меньше (TimeoutCoefficient <= 1)
	a.cfgTimeouts.httpTimeout = time.Duration(a.config.TimeoutCoefficient * to)

	// для проверки просрочки в воркере - сколько времени есть у последнего элемента в очереди
	a.cfgTimeouts.sendExpireTimeout = time.Duration(a.config.TimeoutCoefficient * float64(a.cfgTimeouts.reportInterval))

	// для проверки просрочки при получении списка на отправку
	// dataExpireTimeout должен быть чуть больше sendExpireTimeout - не менее чем на httpTimeout,
	// но не больше reportInterval
	a.cfgTimeouts.dataExpireTimeout = min(
		a.cfgTimeouts.sendExpireTimeout+a.cfgTimeouts.httpTimeout,
		a.cfgTimeouts.reportInterval)

	go a.dataWorker()

	wg := &sync.WaitGroup{}

	//nolint:intrange // for range ReportWorkersCount {...} подсвечивает GoLang inspection
	for i := 0; i < a.config.ReportWorkersCount; i++ {
		wg.Add(1)
		go a.reportWorker(wg)
	}

	wg.Wait()
	close(a.reportChan)

	return nil
}
