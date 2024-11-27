package memstorage

import (
	"sync"
)

type store struct {
	data map[string]map[string]any
	mux  *sync.Mutex
}

func newStore(types ...string) *store {
	s := &store{
		map[string]map[string]any{},
		&sync.Mutex{},
	}
	for _, t := range types {
		s.data[t] = map[string]any{}
	}
	return s
}

func (s store) set(t string, name string, value interface{}) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[t][name] = value
}

func (s store) get(t string, name string) (value any, ok bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	value, ok = s.data[t][name]
	return
}

func (s store) incrInt64(t string, name string, value int64) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if v, ok := s.data[t][name].(int64); ok {
		s.data[t][name] = v + value
	}
}

func (s store) getData() map[string]map[string]any {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.data
}
