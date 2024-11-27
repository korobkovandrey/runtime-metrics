package memstorage

import "sync"

type float64Store struct {
	mux  *sync.Mutex
	data map[string]map[string]float64
}

func newFloat64Store(types ...string) *float64Store {
	s := &float64Store{
		&sync.Mutex{},
		map[string]map[string]float64{},
	}
	for _, t := range types {
		s.data[t] = map[string]float64{}
	}
	return s
}

func (s float64Store) set(t string, name string, value float64) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[t][name] = value
}

func (s float64Store) incr(t string, name string, value float64) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[t][name] += value
}

func (s float64Store) get(t string, name string) (value float64, ok bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	value, ok = s.data[t][name]
	return
}

func (s float64Store) getData() map[string]map[string]float64 {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.data
}
