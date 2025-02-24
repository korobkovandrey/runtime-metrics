package service

import (
	"errors"
	"fmt"
	"slices"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

type Repository interface {
	Find(mr *model.MetricRequest) (*model.Metric, error)
	FindAll() ([]*model.Metric, error)
	Create(mr *model.MetricRequest) (*model.Metric, error)
	Update(mr *model.MetricRequest) (*model.Metric, error)
}

type Service struct {
	r Repository
}

func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

func (s *Service) Update(mr *model.MetricRequest) (*model.Metric, error) {
	m, err := s.r.Find(mr)
	if err != nil && !errors.Is(err, model.ErrMetricNotFound) {
		return m, fmt.Errorf("service.Update: %w", err)
	}
	if m == nil {
		m, err = s.r.Create(mr)
		if err != nil {
			return m, fmt.Errorf("service.Update: %w", err)
		}
		return m, nil
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
		m, err = s.r.Update(mr)
		if err != nil {
			return m, fmt.Errorf("service.Update: %w", err)
		}
		return m, nil
	}
	return m, nil
}

func (s *Service) Find(mr *model.MetricRequest) (*model.Metric, error) {
	m, err := s.r.Find(mr)
	if err != nil {
		return m, fmt.Errorf("service.Find: %w", err)
	}
	return m, nil
}

func (s *Service) FindAll() ([]*model.Metric, error) {
	metrics, err := s.r.FindAll()
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
