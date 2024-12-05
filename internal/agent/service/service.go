package service

import (
	"github.com/korobkovandrey/runtime-metrics/internal/agent/utils"

	"math/rand"
	"strconv"
	"time"
)

const (
	GaugeType   = "gauge"
	CounterType = "counter"
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

func NewGaugeSource() *Source {
	source := &Source{
		gaugeData: map[string]data{},
		collectCount: collectCount{
			sent:   time.Time{},
			expire: time.Now(),
		},
	}
	for i := range utils.GetRuntimeMetrics() {
		source.gaugeData[i] = data{
			sent:   time.Time{}, // time.Now() - для отправки через 10 сек после старта
			expire: time.Now(),
			name:   i,
		}
	}
	source.gaugeData["RandomValue"] = data{
		sent:   time.Time{},
		expire: time.Now(),
		name:   "RandomValue",
	}
	return source
}

func (s *Source) Collect() (err error) {
	result := utils.GetRuntimeMetrics()

	var i, v string

	for i, v = range result {
		if d, ok := s.gaugeData[i]; ok {
			d.value = v
			s.gaugeData[i] = d
		}
	}

	if d, ok := s.gaugeData["RandomValue"]; ok {
		d.value = strconv.FormatFloat(rand.Float64(), 'g', -1, 64)
		s.gaugeData["RandomValue"] = d
	}
	s.collectCount.value++
	return
}

const addCountMetrics = 2

func (s *Source) Len() int {
	return len(utils.GetRuntimeMetrics()) + addCountMetrics
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
				Name:  "PollCount",
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
			if v.Name == "PollCount" {
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
