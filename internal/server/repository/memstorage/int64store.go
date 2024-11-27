package memstorage

import "sync"

type int64Store struct {
	mux  *sync.Mutex
	data map[string]map[string]int64
}

func newInt64Store(types ...string) *int64Store {
	s := &int64Store{
		&sync.Mutex{},
		map[string]map[string]int64{},
	}
	for _, t := range types {
		s.data[t] = map[string]int64{}
	}
	return s
}

func (s int64Store) set(t string, name string, value int64) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[t][name] = value
}

func (s int64Store) incr(t string, name string, value int64) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[t][name] += value
}

func (s int64Store) get(t string, name string) (value int64, ok bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	value, ok = s.data[t][name]
	return
}

func (s int64Store) getData() map[string]map[string]int64 {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.data
}
