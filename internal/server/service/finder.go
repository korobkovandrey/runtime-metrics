package service

import (
	"context"
	"fmt"
	"slices"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

//go:generate mockgen -source=finder.go -destination=mocks/mock_finder.go -package=mocks
type finderRepository interface {
	Find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error)
	FindAll(ctx context.Context) ([]*model.Metric, error)
}

type Finder struct {
	r finderRepository
}

func NewFinder(r finderRepository) *Finder {
	return &Finder{r: r}
}

func (s *Finder) Find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	m, err := s.r.Find(ctx, mr)
	if err != nil {
		return m, fmt.Errorf("failed to find metric: %w", err)
	}
	return m, nil
}

//nolint:dupl // ignore
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
