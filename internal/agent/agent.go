package agent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/service"

	"log"
	"time"
)

type Config struct {
	UpdateGaugeURL         string
	UpdateCounterURL       string
	PollInterval           time.Duration
	ReportInterval         time.Duration
	ReportWorkersCount     int
	HTTPTimeoutCoefficient float64
	httpTimeout            time.Duration
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
	sentChan    chan sentMessage
	reportChan  chan reportMessage
	gaugeSource *service.Source
	config      Config
}

func (a *Agent) makeURL(r reportMessage) string {
	switch r.t {
	case service.GaugeType:
		return a.config.UpdateGaugeURL + r.name + `/` + r.value
	case service.CounterType:
		return a.config.UpdateCounterURL + r.name + `/` + r.value
	default:
		return ``
	}
}

func (a *Agent) pollWorker() {
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

		err = a.gaugeSource.Collect()

		if err != nil {
			log.Println(fmt.Errorf(`pollWorker: %w`, err))
		}

		dataForSend := a.gaugeSource.GetDataForSend(a.config.ReportInterval)

		if len(dataForSend) > 0 {
			// expire для репорта чуть раньше, чем устанавливаемый в GetDataForSend
			expire = time.Now().Add(a.config.ReportInterval - a.config.PollInterval)
			for _, v := range dataForSend {
				a.reportChan <- reportMessage{
					expire: expire,
					t:      v.T,
					name:   v.Name,
					value:  v.Value,
				}
			}
		}

		time.Sleep(a.config.PollInterval)
	}
}

func (a *Agent) reportWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	var (
		now time.Time
		url string
	)
	reader := bytes.NewReader([]byte(``))
	client := &http.Client{
		Timeout: a.config.httpTimeout,
	}
	for c := range a.reportChan {
		if c.expire.Before(time.Now()) {
			continue
		}
		url = a.makeURL(c)
		now = time.Now()
		// @todo сделать в func sendRequest(client *http.Client, reader io.Reader, url string) (ok bool, err error)
		response, err := client.Post(url, `text/plain`, reader)
		if err != nil {
			log.Println(err)
			continue
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Println(err)
		}
		err = response.Body.Close()
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

func New(config Config) *Agent {
	gaugeSource := service.NewGaugeSource()
	gaugeSourceLen := gaugeSource.Len()

	a := &Agent{
		sentChan:    make(chan sentMessage, gaugeSourceLen),
		reportChan:  make(chan reportMessage, gaugeSourceLen),
		gaugeSource: gaugeSource,
		config:      config,
	}

	a.config.httpTimeout = time.Duration(
		a.config.HTTPTimeoutCoefficient * float64(a.config.ReportInterval) /
			float64(a.gaugeSource.Len()) * float64(a.config.ReportWorkersCount),
	)

	return a
}

func (a *Agent) Run() error {
	if a.config.ReportInterval-a.config.PollInterval <= a.config.PollInterval {
		return fmt.Errorf(
			`reportInterval (%ds) must be greater than pollInterval (%ds) by pollInterval (%ds), but %d - %d (%d) > %d`,
			a.config.ReportInterval/time.Second,
			a.config.PollInterval/time.Second,
			a.config.PollInterval/time.Second,
			a.config.ReportInterval/time.Second,
			a.config.PollInterval/time.Second,
			(a.config.ReportInterval-a.config.PollInterval)/time.Second,
			a.config.PollInterval/time.Second)
	}

	go a.pollWorker()

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
