package service

import (
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
	r.EXPECT().FindBatch(gomock.Any(), gomock.Eq([]*model.MetricRequest{
		{Metric: model.NewMetricCounter("testNotExist", 3)},
		{Metric: model.NewMetricCounter("testExist", 13)},
	})).Return([]*model.Metric{model.NewMetricCounter("testExist", 3)}, nil)
	want := []*model.Metric{
		model.NewMetricCounter("testNotExist", 1+2),
		model.NewMetricCounter("testExist", 5+8+3),
		model.NewMetricGauge("testNotExist", 22.5),
		model.NewMetricGauge("testExist", 66.7),
	}
	mrsR := make([]*model.MetricRequest, len(want))
	for i := range want {
		mrsR[i] = &model.MetricRequest{Metric: want[i].Clone()}
	}
	r.EXPECT().CreateOrUpdateBatch(gomock.Any(), gomock.Eq(mrsR)).Return(want, nil)
	got, err := NewBatchUpdater(r).UpdateBatch(t.Context(), []*model.MetricRequest{
		{Metric: model.NewMetricGauge("testNotExist", 12.3)},
		{Metric: model.NewMetricCounter("testNotExist", 1)},
		{Metric: model.NewMetricGauge("testNotExist", 22.5)},
		{Metric: model.NewMetricCounter("testNotExist", 2)},
		{Metric: model.NewMetricGauge("testExist", 55.4)},
		{Metric: model.NewMetricCounter("testExist", 5)},
		{Metric: model.NewMetricGauge("testExist", 66.7)},
		{Metric: model.NewMetricCounter("testExist", 8)},
	})
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
