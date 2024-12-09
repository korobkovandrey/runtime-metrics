package service

import (
	"github.com/korobkovandrey/runtime-metrics/internal/agent/utils"

	"math/rand"
	"strconv"
	"sync"
)

type Source struct {
	mux  *sync.Mutex
	data map[string]string
}

func NewGaugeSource() *Source {
	return &Source{
		mux:  &sync.Mutex{},
		data: map[string]string{},
	}
}

func (s *Source) Collect() {
	s.data["RandomValue"] = strconv.FormatFloat(rand.Float64(), 'g', -1, 64)
	runtimeMetrics := utils.GetRuntimeMetrics()
	var i, v string
	s.mux.Lock()
	defer s.mux.Unlock()
	for i, v = range runtimeMetrics {
		s.data[i] = v
	}
}

func (s *Source) GetDataForSend() (result map[string]string) {
	result = map[string]string{}
	s.mux.Lock()
	defer s.mux.Unlock()
	for i, v := range s.data {
		result[i] = v
	}
	return
}
