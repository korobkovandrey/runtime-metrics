package service

import (
	"math/rand"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/utils"

	"sync"
)

type Source struct {
	mux       *sync.Mutex
	data      map[string]float64
	pollCount int64
}

func NewGaugeSource() *Source {
	return &Source{
		mux:  &sync.Mutex{},
		data: map[string]float64{},
	}
}

func (s *Source) Collect() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data = utils.GetRuntimeMetricsFloat64()
	s.data["RandomValue"] = rand.Float64()
	s.pollCount++
}

func (s *Source) GetDataForSend() (result map[string]float64) {
	result = map[string]float64{}
	s.mux.Lock()
	defer s.mux.Unlock()
	for i, v := range s.data {
		result[i] = v
	}
	return
}

func (s *Source) GetPollCount() int64 {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.pollCount
}
