package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

type MemStorage struct {
	mux   *sync.Mutex
	index map[string]map[string]int
	data  []*model.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		mux:   &sync.Mutex{},
		index: map[string]map[string]int{},
		data:  []*model.Metric{},
	}
}

func (ms *MemStorage) unsafeIndex(mr *model.MetricRequest) (int, bool) {
	i, ok := ms.index[mr.MType][mr.ID]
	return i, ok && i < len(ms.data)
}

func (ms *MemStorage) Find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	if i, ok := ms.unsafeIndex(mr); ok {
		return ms.data[i].Clone(), nil
	}
	return nil, model.ErrMetricNotFound
}

func (ms *MemStorage) unsafeFindAll() []*model.Metric {
	data := make([]*model.Metric, len(ms.data))
	for i := range ms.data {
		data[i] = ms.data[i].Clone()
	}
	return data
}

func (ms *MemStorage) FindAll(ctx context.Context) ([]*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return ms.unsafeFindAll(), nil
}

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

func (ms *MemStorage) Create(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return ms.unsafeCreate(mr)
}

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

func (ms *MemStorage) Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return ms.unsafeUpdate(mr)
}

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

func (ms *MemStorage) CreateOrUpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return ms.unsafeCreateOrUpdateBatch(mrs)
}

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
