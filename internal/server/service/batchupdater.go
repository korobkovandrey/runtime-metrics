package service

import (
	"context"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

//go:generate mockgen -source=batchupdater.go -destination=mocks/mock_batchupdater.go -package=mocks
type batchUpdaterRepository interface {
	FindBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error)
	CreateOrUpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error)
}

type BatchUpdater struct {
	r batchUpdaterRepository
}

func NewBatchUpdater(r batchUpdaterRepository) *BatchUpdater {
	return &BatchUpdater{r: r}
}

//nolint:dupl // ignore
func (s *BatchUpdater) UpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
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
