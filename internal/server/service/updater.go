// Package service contains the service logic.
//
// It provides a way to update metrics.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

// UpdaterRepository is an interface for updating metrics
//
//go:generate mockgen -source=updater.go -destination=mocks/mock_updater.go -package=mocks
type UpdaterRepository interface {
	Find(context.Context, *model.MetricRequest) (*model.Metric, error)
	Create(context.Context, *model.MetricRequest) (*model.Metric, error)
	Update(context.Context, *model.MetricRequest) (*model.Metric, error)
}

// Updater is a service for updating metrics
type Updater struct {
	r UpdaterRepository
}

// NewUpdater returns a new Updater
func NewUpdater(r UpdaterRepository) *Updater {
	return &Updater{r: r}
}

// Update updates the metric
func (s *Updater) Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
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
