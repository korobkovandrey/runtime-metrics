package memstorage

import (
	"slices"
	"sync"
)

type MemStorage struct {
	data map[string]map[string]any
	mux  *sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		map[string]map[string]any{},
		&sync.Mutex{},
	}
}

func (s MemStorage) AddType(t string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.data[t]; ok {
		return
	}
	s.data[t] = map[string]any{}
}

func (s MemStorage) Set(t string, name string, value any) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[t][name] = value
}

func (s MemStorage) IncrInt64(t string, name string, value int64) {
	s.mux.Lock()
	defer s.mux.Unlock()
	v, _ := s.data[t][name].(int64)
	s.data[t][name] = v + value
}

func (s MemStorage) Keys(t string) (result []string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	for i := range s.data[t] {
		result = append(result, i)
	}
	slices.Sort(result)
	return
}

func (s MemStorage) Get(t string, name string) (value any, ok bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	value, ok = s.data[t][name]
	return
}
