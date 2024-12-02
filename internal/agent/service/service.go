package service

import (
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/utils"

	"math/rand"
	"strconv"
	"time"
)

const (
	GaugeType        = `gauge`
	CounterType      = `counter`
	collectCountName = `PollCount`
	randomValueName  = `RandomValue`
)

type data struct {
	sent   time.Time
	expire time.Time
	name   string
	value  string
}

type collectCount struct {
	sent      time.Time
	expire    time.Time
	value     uint64
	sentValue uint64
}

type Source struct {
	gaugeData    map[string]data
	collectCount collectCount
}

var runtimeMetricNames = []string{`Alloc`, `BuckHashSys`, `Frees`, `GCCPUFraction`, `GCSys`, `HeapAlloc`, `HeapIdle`,
	`HeapInuse`, `HeapObjects`, `HeapReleased`, `HeapSys`, `LastGC`, `Lookups`, `MCacheInuse`, `MCacheSys`,
	`MSpanInuse`, `MSpanSys`, `Mallocs`, `NextGC`, `NumForcedGC`, `NumGC`, `OtherSys`, `PauseTotalNs`,
	`StackInuse`, `StackSys`, `Sys`, `TotalAlloc`}

func NewGaugeSource() *Source {
	source := &Source{
		gaugeData: map[string]data{},
		collectCount: collectCount{
			sent:   time.Time{},
			expire: time.Now(),
		},
	}
	for _, i := range runtimeMetricNames {
		source.gaugeData[i] = data{
			sent:   time.Time{}, // time.Now() - для отправки через 10 сек после старта
			expire: time.Now(),
			name:   i,
		}
	}
	source.gaugeData[randomValueName] = data{
		sent:   time.Time{},
		expire: time.Now(),
		name:   randomValueName,
	}
	return source
}

func (s *Source) Collect() (err error) {
	result, errNotNumber, errNotFound := utils.GetRuntimeMetrics(runtimeMetricNames...)

	var i, v string

	for i, v = range result {
		if d, ok := s.gaugeData[i]; ok {
			d.value = v
			s.gaugeData[i] = d
		}
	}

	if d, ok := s.gaugeData[randomValueName]; ok {
		d.value = strconv.FormatFloat(rand.Float64(), 'g', -1, 64)
		s.gaugeData[randomValueName] = d
	}
	s.collectCount.value++

	var errs []any
	errFmt := ``

	if errNotNumber != nil {
		errs = append(errs, errNotNumber)
		errFmt = `%w`
	}
	if errNotFound != nil {
		errs = append(errs, errNotFound)
		if errFmt != `` {
			errFmt += `, `
		}
		errFmt += `%w`
	}
	if len(errs) > 0 {
		err = fmt.Errorf(errFmt, errs...)
	}
	return
}

func (s *Source) Len() int {
	return len(s.gaugeData) + 1
}

type DataForSend struct {
	T     string
	Name  string
	Value string
}

func (s *Source) GetDataForSend(expireTimeout time.Duration, reportInterval time.Duration) (result []DataForSend) {
	now := time.Now()
	expire := now.Add(expireTimeout)
	for i, v := range s.gaugeData {
		// отправляем если:
		// sent + reportInterval < time.Now() - отвечает за отправку раз в интервал
		// и expire < time.Now() - отвечает за отправку в случае таймаута отправки
		if !v.sent.Add(reportInterval).Before(now) {
			continue
		}
		if !v.expire.Before(now) {
			continue
		}
		result = append(result, DataForSend{
			T:     GaugeType,
			Name:  v.name,
			Value: v.value,
		})
		v.expire = expire
		s.gaugeData[i] = v
	}
	if s.collectCount.sent.Add(reportInterval).Before(now) && s.collectCount.expire.Before(now) {
		diffCollectCount := s.getDiffCollectCount()
		if diffCollectCount != 0 {
			s.collectCount.expire = expire
			result = append(result, DataForSend{
				T:     CounterType,
				Name:  collectCountName,
				Value: strconv.Itoa(diffCollectCount),
			})
		}
	}
	return
}

type DataSent struct {
	Sent  time.Time
	T     string
	Name  string
	Value string
}

func (s *Source) SetDataSent(sent []DataSent) {
	for _, v := range sent {
		if v.T == CounterType {
			if v.Name == collectCountName {
				s.collectCount.sent = v.Sent
				s.addCollectCountSent(v.Value)
			}
			continue
		}
		d, ok := s.gaugeData[v.Name]
		if !ok {
			continue
		}
		d.sent = v.Sent
		s.gaugeData[v.Name] = d
	}
}

func (s *Source) getDiffCollectCount() int {
	return int(s.collectCount.value - s.collectCount.sentValue)
}

func (s *Source) addCollectCountSent(addValue string) {
	value, err := strconv.Atoi(addValue)
	if err == nil {
		s.collectCount.sentValue += uint64(value)
	}
}
