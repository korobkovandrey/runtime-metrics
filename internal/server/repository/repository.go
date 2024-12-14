package repository

import (
	"errors"
	"fmt"
	"slices"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/adapter"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository/memstorage"
)

type Adapter interface {
	Update(name string, value string) error
	GetStorageValue(name string) (any, bool)
	Names() []string
}

type Store struct {
	repository adapter.Repository
	data       map[string]Adapter
}

var (
	ErrTypeIsNotValid  = errors.New("type is not valid")
	ErrValueIsRequired = errors.New("value is required")
)

func (s Store) Get(t string) (Adapter, error) {
	if s.data[t] == nil {
		return nil, fmt.Errorf(`"%v" %w`, t, ErrTypeIsNotValid)
	}
	return s.data[t], nil
}

const (
	gaugeType   = "gauge"
	counterType = "counter"
)

func (s Store) UpdateMetrics(metrics *model.Metrics) error {
	switch metrics.MType {
	case gaugeType:
		if metrics.Value == nil {
			return fmt.Errorf("UpdateMetrics: %w", ErrValueIsRequired)
		}
		s.repository.SetFloat64(metrics.MType, metrics.ID, *metrics.Value)
		if v, ok := s.repository.GetFloat64(metrics.MType, metrics.ID); ok {
			*metrics.Value = v
		} else {
			metrics.Value = nil
		}
	case counterType:
		if metrics.Delta == nil {
			return fmt.Errorf("UpdateMetrics: %w", ErrValueIsRequired)
		}
		s.repository.IncrInt64(metrics.MType, metrics.ID, *metrics.Delta)
		if v, ok := s.repository.GetInt64(metrics.MType, metrics.ID); ok {
			*metrics.Delta = v
		} else {
			metrics.Delta = nil
		}
	default:
		return fmt.Errorf(`UpdateMetrics: "%v" %w`, metrics.MType, ErrTypeIsNotValid)
	}
	return nil
}

func (s Store) FillMetrics(metrics *model.Metrics) error {
	switch metrics.MType {
	case gaugeType:
		if v, ok := s.repository.GetFloat64(metrics.MType, metrics.ID); ok {
			metrics.Value = &v
		} else {
			metrics.Value = nil
		}
	case counterType:
		if v, ok := s.repository.GetInt64(metrics.MType, metrics.ID); ok {
			metrics.Delta = &v
		} else {
			metrics.Delta = nil
		}
	default:
		return fmt.Errorf(`UpdateMetrics: "%v" %w`, metrics.MType, ErrTypeIsNotValid)
	}
	return nil
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

func NewStore(repository adapter.Repository) *Store {
	store := &Store{
		repository,
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
