// Package repository MemStorage is a simple in-memory storage for metrics.
//
// It supports the following operations:
//
// - Create: adds a new metric to the storage.
// - Update: updates an existing metric in the storage.
// - CreateOrUpdateBatch: adds or updates multiple metrics in the storage.
// - Find: finds a metric in the storage by its ID and name.
// - FindAll: finds all metrics in the storage.
//
// The storage is thread-safe and provides a simple locking mechanism.
package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

// MemStorage is a simple in-memory storage for metrics.
type MemStorage struct {
	mux   *sync.Mutex
	index map[string]map[string]int
	data  []*model.Metric
}

// NewMemStorage creates a new MemStorage.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		mux:   &sync.Mutex{},
		index: map[string]map[string]int{},
		data:  []*model.Metric{},
	}
}

// unsafeIndex returns the index of the metric in the data slice.
func (ms *MemStorage) unsafeIndex(mr *model.MetricRequest) (int, bool) {
	i, ok := ms.index[mr.MType][mr.ID]
	return i, ok && i < len(ms.data)
}

// Find returns the metric with the given ID.
func (ms *MemStorage) Find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	if i, ok := ms.unsafeIndex(mr); ok {
		return ms.data[i].Clone(), nil
	}
	return nil, model.ErrMetricNotFound
}

// unsafeFindAll returns all metrics.
func (ms *MemStorage) unsafeFindAll() []*model.Metric {
	data := make([]*model.Metric, len(ms.data))
	for i := range ms.data {
		data[i] = ms.data[i].Clone()
	}
	return data
}

// FindAll returns all metrics.
func (ms *MemStorage) FindAll(ctx context.Context) ([]*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return ms.unsafeFindAll(), nil
}

// FindBatch returns the metrics with the given IDs.
func (ms *MemStorage) FindBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	var res []*model.Metric
	for _, mr := range mrs {
		if i, ok := ms.unsafeIndex(mr); ok {
			res = append(res, ms.data[i].Clone())
		}
	}
	return res, nil
}

// unsafeCreate creates a new metric.
func (ms *MemStorage) unsafeCreate(mr *model.MetricRequest) (*model.Metric, error) {
	if _, ok := ms.unsafeIndex(mr); ok {
		return nil, model.ErrMetricAlreadyExist
	}
	if _, ok := ms.index[mr.MType]; !ok {
		ms.index[mr.MType] = map[string]int{}
	}
	ms.index[mr.MType][mr.ID] = len(ms.data)
	ms.data = append(ms.data, mr.Clone())
	return ms.data[ms.index[mr.MType][mr.ID]].Clone(), nil
}

// Create creates a new metric.
func (ms *MemStorage) Create(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return ms.unsafeCreate(mr)
}

// unsafeUpdate updates the metric.
func (ms *MemStorage) unsafeUpdate(mr *model.MetricRequest) (*model.Metric, error) {
	i, ok := ms.unsafeIndex(mr)
	if !ok {
		return nil, model.ErrMetricNotFound
	}
	if mr.Value != nil {
		if ms.data[i].Value == nil {
			ms.data[i].Value = new(float64)
		}
		*ms.data[i].Value = *mr.Value
	}
	if mr.Delta != nil {
		if ms.data[i].Delta == nil {
			ms.data[i].Delta = new(int64)
		}
		*ms.data[i].Delta = *mr.Delta
	}
	return ms.data[i].Clone(), nil
}

// Update updates the metric.
func (ms *MemStorage) Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return ms.unsafeUpdate(mr)
}

// unsafeCreateOrUpdateBatch creates or updates the metrics.
func (ms *MemStorage) unsafeCreateOrUpdateBatch(mrs []*model.MetricRequest) ([]*model.Metric, error) {
	res := make([]*model.Metric, len(mrs))
	for i, mr := range mrs {
		var m *model.Metric
		var err error
		if _, ok := ms.unsafeIndex(mr); ok {
			m, err = ms.unsafeUpdate(mr)
			if err != nil {
				return nil, fmt.Errorf("failed to update metric: %w", err)
			}
		} else {
			m, err = ms.unsafeCreate(mr)
			if err != nil {
				return nil, fmt.Errorf("failed to create metric: %w", err)
			}
		}
		res[i] = m
	}
	return res, nil
}

// CreateOrUpdateBatch creates or updates the metrics.
func (ms *MemStorage) CreateOrUpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return ms.unsafeCreateOrUpdateBatch(mrs)
}

// fill fills the storage with the given metrics.
func (ms *MemStorage) fill(data []*model.Metric) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	ms.index = map[string]map[string]int{}
	ms.data = make([]*model.Metric, len(data))
	for i := range data {
		if _, ok := ms.index[data[i].MType]; !ok {
			ms.index[data[i].MType] = map[string]int{}
		}
		ms.index[data[i].MType][data[i].ID] = i
		ms.data[i] = data[i].Clone()
	}
}
