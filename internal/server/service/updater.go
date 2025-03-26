package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

//go:generate mockgen -source=updater.go -destination=mocks/mock_updater.go -package=mocks
type updaterRepository interface {
	Find(context.Context, *model.MetricRequest) (*model.Metric, error)
	Create(context.Context, *model.MetricRequest) (*model.Metric, error)
	Update(context.Context, *model.MetricRequest) (*model.Metric, error)
}

type Updater struct {
	r updaterRepository
}

func NewUpdater(r updaterRepository) *Updater {
	return &Updater{r: r}
}

func (s *Updater) Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	fmt.Println(mr)
	m, err := s.r.Find(ctx, mr)
	if err != nil {
		if !errors.Is(err, model.ErrMetricNotFound) {
			return m, fmt.Errorf("failed to find metric: %w", err)
		}
		m, err = s.r.Create(ctx, mr)
		if err == nil {
			return m, nil
		}
		if !errors.Is(err, model.ErrMetricAlreadyExist) {
			return m, fmt.Errorf("failed to create metric: %w", err)
		}
		m, err = s.r.Find(ctx, mr)
		if err != nil {
			return m, fmt.Errorf("failed to find metric: %w", err)
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
			return m, fmt.Errorf("failed to update metric: %w", err)
		}
		return m, nil
	}
	return m, nil
}
