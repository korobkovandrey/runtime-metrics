package repository

import (
	"errors"
	"fmt"
	"slices"

	"github.com/korobkovandrey/runtime-metrics/internal/server/adapter"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository/memstorage"
)

type Adapter interface {
	Update(name string, value string) error
	GetStorageValue(name string) (any, bool)
	Names() []string
}

type Store struct {
	data map[string]Adapter
}

var (
	ErrTypeIsNotValid = errors.New("type is not valid")
)

func (s Store) Get(t string) (Adapter, error) {
	if s.data[t] == nil {
		return nil, fmt.Errorf(`"%v" %w`, t, ErrTypeIsNotValid)
	}
	return s.data[t], nil
}

type StorageValue struct {
	Value any
	T     string
	Name  string
}

func (s Store) GetAllData() (result []StorageValue) {
	types := make([]string, 0, len(s.data))
	for i := range s.data {
		types = append(types, i)
	}
	slices.Sort(types)
	for _, i := range types {
		for _, k := range s.data[i].Names() {
			v, ok := s.data[i].GetStorageValue(k)
			if !ok {
				continue
			}
			result = append(result, StorageValue{
				T:     i,
				Name:  k,
				Value: v,
			})
		}
	}
	return
}

func (s Store) addAdapter(key string, a Adapter) {
	s.data[key] = a
}

const (
	gaugeType   = "gauge"
	counterType = "counter"
)

func NewStore(repository adapter.Repository) *Store {
	store := &Store{
		map[string]Adapter{},
	}
	repository.AddType(gaugeType)
	store.addAdapter(gaugeType, adapter.NewGauge(repository, gaugeType))
	repository.AddType(counterType)
	store.addAdapter(counterType, adapter.NewCounter(repository, counterType))
	return store
}

func NewStoreMemStorage() *Store {
	return NewStore(memstorage.NewMemStorage())
}
