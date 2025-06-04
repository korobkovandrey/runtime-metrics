// Package service contains the service logic.
//
// The service provides methods for working with metrics.
package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

// BatchUpdaterRepository is an interface for batch updating metrics
//
//go:generate mockgen -source=batchupdater.go -destination=mocks/mock_batchupdater.go -package=mocks
type BatchUpdaterRepository interface {
	FindBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error)
	CreateOrUpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error)
}

// BatchUpdater is a service for batch updating metrics.
type BatchUpdater struct {
	r BatchUpdaterRepository
}

// NewBatchUpdater returns a service for batch updating metrics.
func NewBatchUpdater(r BatchUpdaterRepository) *BatchUpdater {
	return &BatchUpdater{r: r}
}

// UpdateBatch updates the metrics.
func (s *BatchUpdater) UpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
	var mrsReq []*model.MetricRequest
	mrsGaugeIndexMap := map[string]int{}
	mrsCounterMap := map[string]*model.MetricRequest{}
	for i := range mrs {
		if mrs[i].MType != model.TypeCounter {
			mrsGaugeIndexMap[mrs[i].ID] = i
			continue
		}
		if _, ok := mrsCounterMap[mrs[i].ID]; ok {
			*mrsCounterMap[mrs[i].ID].Delta += *mrs[i].Delta
		} else {
			mrsCounterMap[mrs[i].ID] = mrs[i]
			mrsReq = append(mrsReq, mrsCounterMap[mrs[i].ID])
		}
	}
	if len(mrsReq) > 0 {
		mCounterExist, err := s.r.FindBatch(ctx, mrsReq)
		if err != nil {
			return nil, fmt.Errorf("failed to find batch: %w", err)
		}
		for i := range mCounterExist {
			*mrsCounterMap[mCounterExist[i].ID].Delta += *mCounterExist[i].Delta
		}
	}

	if len(mrsGaugeIndexMap) > 0 {
		indexes := make([]int, len(mrsGaugeIndexMap))
		k := 0
		for _, i := range mrsGaugeIndexMap {
			indexes[k] = i
			k++
		}
		sort.Ints(indexes)
		for _, i := range indexes {
			mrsReq = append(mrsReq, mrs[i])
		}
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
