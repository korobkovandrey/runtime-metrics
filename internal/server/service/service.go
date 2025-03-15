package service

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

//go:generate mockgen -source=service.go -destination=../mocks/service.go -package=mocks

type Repository interface {
	Find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error)
	FindAll(ctx context.Context) ([]*model.Metric, error)
	FindBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error)
	Create(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error)
	Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error)
	CreateOrUpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error)
}

type Service struct {
	r Repository
}

func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

func (s *Service) Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	const errorMsg = "service.Update: %w"
	m, err := s.r.Find(ctx, mr)
	if err != nil {
		if !errors.Is(err, model.ErrMetricNotFound) {
			return m, fmt.Errorf(errorMsg, err)
		}
		m, err = s.r.Create(ctx, mr)
		if err == nil {
			return m, nil
		}
		if !errors.Is(err, model.ErrMetricAlreadyExist) {
			return m, fmt.Errorf(errorMsg, err)
		}
		m, err = s.r.Find(ctx, mr)
		if err != nil {
			return m, fmt.Errorf(errorMsg, err)
		}
	}

	needUpdate := false
	switch mr.MType {
	case model.TypeGauge:
		if mr.Value != nil {
			needUpdate = true
		}
	case model.TypeCounter:
		if mr.Delta != nil {
			if m.Delta != nil {
				*mr.Delta += *m.Delta
			}
			needUpdate = true
		}
	}
	if needUpdate {
		m, err = s.r.Update(ctx, mr)
		if err != nil {
			return m, fmt.Errorf(errorMsg, err)
		}
		return m, nil
	}
	return m, nil
}

func (s *Service) UpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
	var mrsReq []*model.MetricRequest
	mrsGaugeIndexMap := map[string]int{}
	mrsCounterMap := map[string]*model.MetricRequest{}
	for i, mr := range mrs {
		if mr.MType != model.TypeCounter {
			mrsGaugeIndexMap[mr.ID] = i
			continue
		}
		if _, ok := mrsCounterMap[mr.ID]; ok {
			*mrsCounterMap[mr.ID].Delta += *mr.Delta
		} else {
			mrsCounterMap[mr.ID] = mr
			mrsReq = append(mrsReq, mr)
		}
	}
	if len(mrsReq) > 0 {
		mCounterExist, err := s.r.FindBatch(ctx, mrsReq)
		if err != nil {
			return nil, fmt.Errorf("failed to find batch: %w", err)
		}
		for _, mr := range mCounterExist {
			*mrsCounterMap[mr.ID].Delta += *mr.Delta
		}
	}
	for _, i := range mrsGaugeIndexMap {
		mrsReq = append(mrsReq, mrs[i])
	}
	if len(mrsReq) == 0 {
		return []*model.Metric{}, nil
	}
	res, err := s.r.CreateOrUpdateBatch(ctx, mrsReq)
	if err != nil {
		return res, fmt.Errorf("failed to update batch: %w", err)
	}
	return res, nil
}

func (s *Service) Find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	m, err := s.r.Find(ctx, mr)
	if err != nil {
		return m, fmt.Errorf("service.Find: %w", err)
	}
	return m, nil
}

func (s *Service) FindAll(ctx context.Context) ([]*model.Metric, error) {
	metrics, err := s.r.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service.FindAll: %w", err)
	}
	slices.SortFunc(metrics, func(a *model.Metric, b *model.Metric) int {
		if a.MType == b.MType {
			if a.ID > b.ID {
				return 1
			} else if a.ID < b.ID {
				return -1
			}
			return 0
		}
		if a.MType > b.MType {
			return 1
		}
		return -1
	})
	return metrics, nil
}
