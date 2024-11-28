package repository

import (
	"errors"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/adapter"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository/memstorage"
)

type Adapter interface {
	Update(name string, value string) error
	GetStorageData() interface{}
}

type Store struct {
	data map[string]Adapter
}

var (
	ErrTypeIsNotValid = errors.New(`type is not valid`)
)

func (s Store) Get(t string) (Adapter, error) {
	if s.data[t] == nil {
		return nil, fmt.Errorf(`"%v" %w`, t, ErrTypeIsNotValid)
	}
	return s.data[t], nil
}

func (s Store) addAdapter(key string, a Adapter) {
	s.data[key] = a
}

const (
	gaugeType   = `gauge`
	counterType = `counter`
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
