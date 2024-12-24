package repository

import (
	"errors"
	"fmt"
	"slices"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/adapter"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository/memstorage"
)

type Adapter interface {
	Update(name string, value string) error
	GetStorageValue(name string) (any, bool)
	Names() []string
}

type Store struct {
	config     *config.Config
	repository adapter.Repository
	data       map[string]Adapter
}

var (
	ErrTypeIsNotValid  = errors.New("type is not valid")
	ErrValueIsRequired = errors.New("value is required")
	ErrMetricNotFound  = errors.New("metric not found")
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

func (s Store) UpdateMetric(metric *model.Metric) error {
	switch metric.MType {
	case gaugeType:
		if metric.Value == nil {
			return fmt.Errorf("UpdateMetric: %w", ErrValueIsRequired)
		}
		s.repository.SetFloat64(metric.MType, metric.ID, *metric.Value)
		if v, ok := s.repository.GetFloat64(metric.MType, metric.ID); ok {
			*metric.Value = v
		} else {
			metric.Value = nil
		}
	case counterType:
		if metric.Delta == nil {
			return fmt.Errorf("UpdateMetric: %w", ErrValueIsRequired)
		}
		s.repository.IncrInt64(metric.MType, metric.ID, *metric.Delta)
		if v, ok := s.repository.GetInt64(metric.MType, metric.ID); ok {
			*metric.Delta = v
		} else {
			metric.Delta = nil
		}
	default:
		return fmt.Errorf(`UpdateMetric: "%v" %w`, metric.MType, ErrTypeIsNotValid)
	}
	return nil
}

func (s Store) FillMetric(metric *model.Metric) error {
	switch metric.MType {
	case gaugeType:
		if v, ok := s.repository.GetFloat64(metric.MType, metric.ID); ok {
			metric.Value = &v
		} else {
			return fmt.Errorf(`FillMetric %s.%s: %w`, metric.MType, metric.ID, ErrMetricNotFound)
		}
	case counterType:
		if v, ok := s.repository.GetInt64(metric.MType, metric.ID); ok {
			metric.Delta = &v
		} else {
			return fmt.Errorf(`FillMetric %s.%s: %w`, metric.MType, metric.ID, ErrMetricNotFound)
		}
	default:
		return fmt.Errorf(`FillMetric %s: %w`, metric.MType, ErrTypeIsNotValid)
	}
	return nil
}

type StorageValue struct {
	Value any    `json:"value"`
	T     string `json:"type"`
	Name  string `json:"name"`
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

func NewStore(cfg *config.Config, repository adapter.Repository) *Store {
	store := &Store{
		config:     cfg,
		repository: repository,
		data:       map[string]Adapter{},
	}
	repository.AddType(gaugeType)
	store.addAdapter(gaugeType, adapter.NewGauge(repository, gaugeType))
	repository.AddType(counterType)
	store.addAdapter(counterType, adapter.NewCounter(repository, counterType))
	return store
}

func NewStoreMemStorage(cfg *config.Config) (*Store, error) {
	store := NewStore(cfg, memstorage.NewMemStorage())
	if store.config.Restore {
		err := store.restore()
		if err != nil {
			return nil, fmt.Errorf("NewStoreMemStorage: %w", err)
		}
	}
	go store.Run()
	return store, nil
}
