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

type Config struct {
	UpdateGaugeURL     string
	UpdateCounterURL   string
	PollInterval       time.Duration
	ReportInterval     time.Duration
	ReportWorkersCount int
	TimeoutCoefficient float64
	httpTimeout        time.Duration
	sendExpireTimeout  time.Duration
	dataExpireTimeout  time.Duration
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
	config        Config
}

func (a *Agent) makeURL(r reportMessage) (string, error) {
	switch r.t {
	case service.GaugeType:
		return a.config.UpdateGaugeURL + r.name + `/` + r.value, nil
	case service.CounterType:
		return a.config.UpdateCounterURL + r.name + `/` + r.value, nil
	default:
		return ``, fmt.Errorf(`%s: %w`, r.t, errors.New(`type is not valid`))
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

		// получение данных раз PollInterval
		if a.collectExpire.Before(time.Now()) {
			a.collectExpire = time.Now().Add(a.config.PollInterval)
			err = a.gaugeSource.Collect()
			if err != nil {
				log.Println(fmt.Errorf(`dataWorker: %w`, err))
			}
		}

		// метрики, которые не отправлялись дольше ReportInterval
		dataForSend := a.gaugeSource.GetDataForSend(a.config.dataExpireTimeout, a.config.ReportInterval)

		if len(dataForSend) > 0 {
			expire = time.Now().Add(a.config.sendExpireTimeout)
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
		Timeout: a.config.httpTimeout,
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
		response, err := client.Post(url, `text/plain`, http.NoBody)
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

func New(cfg config.Config) *Agent {
	gaugeSource := service.NewGaugeSource()
	gaugeSourceLen := gaugeSource.Len()

	a := &Agent{
		sentChan:    make(chan sentMessage, gaugeSourceLen),
		reportChan:  make(chan reportMessage, gaugeSourceLen),
		gaugeSource: gaugeSource,
		config: Config{
			UpdateGaugeURL:     cfg.UpdateGaugeURL,
			UpdateCounterURL:   cfg.UpdateCounterURL,
			PollInterval:       cfg.PollInterval,
			ReportInterval:     cfg.ReportInterval,
			ReportWorkersCount: cfg.ReportWorkersCount,
			TimeoutCoefficient: cfg.TimeoutCoefficient,
		},
	}

	return a
}

func (a *Agent) Run() error {
	if a.config.ReportInterval <= a.config.PollInterval {
		return fmt.Errorf(`ReportInterval (%ds) must be greater than PollInterval (%ds)`,
			a.config.ReportInterval/time.Second, a.config.PollInterval/time.Second)
	}
	if a.config.TimeoutCoefficient > 1 {
		return fmt.Errorf(`TimeoutCoefficient (%f) must be less than 1 `, a.config.TimeoutCoefficient)
	}

	// максимальное время на запрос, чтобы успеть уложить отправку всех метрик в ReportInterval
	to := float64(a.config.ReportInterval) / float64(a.gaugeSource.Len()) * float64(a.config.ReportWorkersCount)

	// таймаут для http клиента делаем меньше (TimeoutCoefficient <= 1)
	a.config.httpTimeout = time.Duration(a.config.TimeoutCoefficient * to)

	// для проверки просрочки в воркере - сколько времени есть у последнего элемента в очереди
	a.config.sendExpireTimeout = time.Duration(a.config.TimeoutCoefficient * float64(a.config.ReportInterval))

	// для проверки просрочки при получении списка на отправку
	// dataExpireTimeout должен быть чуть больше sendExpireTimeout - не менее чем на httpTimeout,
	// но не больше ReportInterval
	a.config.dataExpireTimeout = min(a.config.sendExpireTimeout+a.config.httpTimeout, a.config.ReportInterval)

	go a.dataWorker()

	wg := &sync.WaitGroup{}

	//nolint:intrange // for range reportWorkersCount {...} подсвечивает GoLang inspection
	for i := 0; i < a.config.ReportWorkersCount; i++ {
		wg.Add(1)
		go a.reportWorker(wg)
	}

	wg.Wait()
	close(a.reportChan)

	return nil
}
