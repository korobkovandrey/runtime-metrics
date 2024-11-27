package repository

import (
	"errors"
	"fmt"
	"github.com/korobkovandrey/runtime-metrics/internal/server/adapter"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository/memstorage"
)

type metricType string

const (
	gaugeType   metricType = `gauge`
	counterType metricType = `counter`
)

type Adapter interface {
	Update(name string, value string) error
	GetStorageData() interface{}
}

type Store struct {
	data map[metricType]Adapter
}

var (
	ErrTypeIsNotValid = errors.New(`type is not valid`)
)

func (s Store) Get(strType string) (Adapter, error) {
	t := metricType(strType)
	if s.data[t] == nil {
		return nil, fmt.Errorf(`"%v" %w`, strType, ErrTypeIsNotValid)
	}
	return s.data[t], nil
}

func NewStoreMemStorage() *Store {
	memStorage := memstorage.NewMemStorage()
	store := &Store{
		map[metricType]Adapter{
			gaugeType:   adapter.NewGauge(memStorage),
			counterType: adapter.NewCounter(memStorage),
		},
	}
	return store
}
