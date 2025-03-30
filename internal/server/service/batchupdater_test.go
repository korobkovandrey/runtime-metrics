package service

import (
	"context"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBatchUpdater_UpdateBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	r := mocks.NewMockBatchUpdaterRepository(ctrl)
	s := NewBatchUpdater(r)
	r.EXPECT().FindBatch(gomock.Any(), gomock.Eq([]*model.MetricRequest{
		{Metric: model.NewMetricCounter("testNotExist", 3)},
		{Metric: model.NewMetricCounter("testExist", 13)},
	})).Return([]*model.Metric{model.NewMetricCounter("testExist", 3)}, nil)
	mrsReq := []*model.MetricRequest{
		{Metric: model.NewMetricCounter("testNotExist", 1+2)},
		{Metric: model.NewMetricCounter("testExist", 5+8+3)},
		{Metric: model.NewMetricGauge("testNotExist", 22.55)},
		{Metric: model.NewMetricGauge("testExist", 66.7)},
	}
	var want []*model.Metric
	for _, mr := range mrsReq {
		want = append(want, mr.Clone())
	}
	r.EXPECT().CreateOrUpdateBatch(gomock.Any(), gomock.Eq(mrsReq)).Return(want, nil)
	got, err := s.UpdateBatch(context.TODO(), []*model.MetricRequest{
		{Metric: model.NewMetricGauge("testNotExist", 12.34)},
		{Metric: model.NewMetricCounter("testNotExist", 1)},
		{Metric: model.NewMetricGauge("testNotExist", 22.55)},
		{Metric: model.NewMetricCounter("testNotExist", 2)},
		{Metric: model.NewMetricGauge("testExist", 55.44)},
		{Metric: model.NewMetricCounter("testExist", 5)},
		{Metric: model.NewMetricGauge("testExist", 66.7)},
		{Metric: model.NewMetricCounter("testExist", 8)},
	})
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
