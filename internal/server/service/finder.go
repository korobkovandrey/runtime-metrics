// Package service Finder is a structure that provides a way to find metrics
//
// It contains a repository, which is used to find metrics. The Finder
// structure is a simple wrapper around the repository, which provides
// a convenient way to find metrics.
package service

import (
	"context"
	"fmt"
	"slices"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

// FinderRepository is a repository for finding metrics
//
//go:generate mockgen -source=finder.go -destination=mocks/mock_finder.go -package=mocks
type FinderRepository interface {
	Find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error)
	FindAll(ctx context.Context) ([]*model.Metric, error)
}

// Finder is a structure that provides a way to find metrics
type Finder struct {
	r FinderRepository
}

// NewFinder returns a new finder
func NewFinder(r FinderRepository) *Finder {
	return &Finder{r: r}
}

// Find returns the metric with the given ID
func (s *Finder) Find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	m, err := s.r.Find(ctx, mr)
	if err != nil {
		return m, fmt.Errorf("failed to find metric: %w", err)
	}
	return m, nil
}

// FindAll returns all metrics
func (s *Finder) FindAll(ctx context.Context) ([]*model.Metric, error) {
	metrics, err := s.r.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all metrics: %w", err)
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
