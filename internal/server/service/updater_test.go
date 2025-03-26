package service

import (
	"context"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUpdater_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("creating", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "1")
		require.NoError(t, err)
		want := mr.Clone()
		ctx := context.TODO()
		r := mocks.NewMockupdaterRepository(ctrl)
		r.EXPECT().Find(gomock.Eq(ctx), gomock.Eq(mr)).Return(nil, model.ErrMetricNotFound)
		r.EXPECT().Create(gomock.Eq(ctx), gomock.Eq(mr)).Return(want, nil)
		r.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
		s := NewUpdater(r)
		got, err := s.Update(ctx, mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("updating counter", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeCounter, "test", "10")
		require.NoError(t, err)
		memMetric := model.NewMetricCounter("test", 1)
		want := memMetric.Clone()
		*want.Delta += *mr.Delta
		ctx := context.TODO()
		r := mocks.NewMockupdaterRepository(ctrl)
		r.EXPECT().Find(gomock.Eq(ctx), gomock.Eq(mr)).Return(memMetric, nil)
		r.EXPECT().Create(gomock.Any(), gomock.Any()).MaxTimes(0)
		r.EXPECT().Update(gomock.Eq(ctx), gomock.Eq(&model.MetricRequest{Metric: want})).Return(want, nil)
		s := NewUpdater(r)
		got, err := s.Update(ctx, mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})
}
