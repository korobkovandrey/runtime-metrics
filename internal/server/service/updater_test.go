package service

import (
	"context"
	"errors"
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
		r := mocks.NewMockUpdaterRepository(ctrl)
		r.EXPECT().Find(gomock.Any(), gomock.Eq(mr)).Return(nil, model.ErrMetricNotFound)
		r.EXPECT().Create(gomock.Any(), gomock.Eq(mr)).Return(want, nil)
		r.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
		s := NewUpdater(r)
		got, err := s.Update(context.TODO(), mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("updating counter", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeCounter, "test", "10")
		require.NoError(t, err)
		memMetric := model.NewMetricCounter("test", 1)
		want := memMetric.Clone()
		*want.Delta += *mr.Delta
		r := mocks.NewMockUpdaterRepository(ctrl)
		r.EXPECT().Find(gomock.Any(), gomock.Eq(mr)).Return(memMetric, nil)
		r.EXPECT().Create(gomock.Any(), gomock.Any()).MaxTimes(0)
		r.EXPECT().Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: want})).Return(want, nil)
		s := NewUpdater(r)
		got, err := s.Update(context.TODO(), mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("updating gauge", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "10")
		require.NoError(t, err)
		memMetric := model.NewMetricGauge("test", 1)
		want := memMetric.Clone()
		*want.Value = *mr.Value
		r := mocks.NewMockUpdaterRepository(ctrl)
		r.EXPECT().Find(gomock.Any(), gomock.Eq(mr)).Return(memMetric, nil)
		r.EXPECT().Create(gomock.Any(), gomock.Any()).MaxTimes(0)
		r.EXPECT().Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: want})).Return(want, nil)
		s := NewUpdater(r)
		got, err := s.Update(context.TODO(), mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("error", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "10")
		require.NoError(t, err)
		r := mocks.NewMockUpdaterRepository(ctrl)
		r.EXPECT().Find(gomock.Any(), gomock.Eq(mr)).Return(nil, errors.New("error"))
		r.EXPECT().Create(gomock.Any(), gomock.Any()).MaxTimes(0)
		r.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
		s := NewUpdater(r)
		got, err := s.Update(context.TODO(), mr)
		assert.Nil(t, got)
		assert.Error(t, err)
	})

	t.Run("error creating", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "10")
		require.NoError(t, err)
		r := mocks.NewMockUpdaterRepository(ctrl)
		r.EXPECT().Find(gomock.Any(), gomock.Eq(mr)).Return(nil, model.ErrMetricNotFound)
		r.EXPECT().Create(gomock.Any(), gomock.Eq(mr)).Return(nil, errors.New("error"))
		r.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
		s := NewUpdater(r)
		got, err := s.Update(context.TODO(), mr)
		assert.Nil(t, got)
		assert.Error(t, err)
	})

	t.Run("error creating already exist", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "10")
		require.NoError(t, err)
		r := mocks.NewMockUpdaterRepository(ctrl)
		r.EXPECT().Find(gomock.Any(), gomock.Eq(mr)).Times(2).Return(nil, model.ErrMetricNotFound)
		r.EXPECT().Create(gomock.Any(), gomock.Eq(mr)).Return(nil, model.ErrMetricAlreadyExist)
		r.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
		s := NewUpdater(r)
		got, err := s.Update(context.TODO(), mr)
		assert.Nil(t, got)
		assert.Error(t, err)
	})

	t.Run("error updating", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "10")
		require.NoError(t, err)
		memMetric := model.NewMetricGauge("test", 1)
		want := memMetric.Clone()
		*want.Value = *mr.Value
		r := mocks.NewMockUpdaterRepository(ctrl)
		r.EXPECT().Find(gomock.Any(), gomock.Eq(mr)).Return(memMetric, nil)
		r.EXPECT().Create(gomock.Any(), gomock.Any()).MaxTimes(0)
		r.EXPECT().Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: want})).Return(nil, errors.New("error"))
		s := NewUpdater(r)
		got, err := s.Update(context.TODO(), mr)
		assert.Nil(t, got)
		assert.Error(t, err)
	})
}
