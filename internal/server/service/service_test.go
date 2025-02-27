package service

import (
	"errors"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepository := mocks.NewMockRepository(ctrl)
	s := NewService(mockRepository)
	assert.Equal(t, s, &Service{mockRepository})
}

func TestService_Find(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepository := mocks.NewMockRepository(ctrl)

	t.Run("valid", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "1")
		require.NoError(t, err)
		want := mr.Clone()
		mockRepository.EXPECT().Find(gomock.Eq(mr)).Return(want, nil)
		service := NewService(mockRepository)
		got, err := service.Find(mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("error", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "1")
		require.NoError(t, err)
		mockRepository.EXPECT().Find(gomock.Eq(mr)).Return(nil, model.ErrMetricNotFound)
		service := NewService(mockRepository)
		got, err := service.Find(mr)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, model.ErrMetricNotFound)
	})
}

func TestService_FindAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepository := mocks.NewMockRepository(ctrl)
	t.Run("valid", func(t *testing.T) {
		want := []*model.Metric{
			model.NewMetricGauge("test1", 1),
			model.NewMetricCounter("test2", 1),
		}
		mockRepository.EXPECT().FindAll().Return(want, nil)
		service := NewService(mockRepository)
		got, err := service.FindAll()
		assert.NoError(t, err)
		assert.ElementsMatch(t, want, got)
	})

	t.Run("error", func(t *testing.T) {
		mockRepository.EXPECT().FindAll().
			Return([]*model.Metric{}, errors.New("error"))
		service := NewService(mockRepository)
		got, err := service.FindAll()
		assert.Nil(t, got)
		assert.Error(t, err)
	})
}

func TestService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepository := mocks.NewMockRepository(ctrl)

	t.Run("creating", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "1")
		require.NoError(t, err)
		want := mr.Clone()
		mockRepository.EXPECT().Find(gomock.Eq(mr)).Return(nil, model.ErrMetricNotFound)
		mockRepository.EXPECT().Create(gomock.Eq(mr)).Return(want, nil)
		mockRepository.EXPECT().Update(gomock.Any()).MaxTimes(0)
		service := NewService(mockRepository)
		got, err := service.Update(mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("updating counter", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeCounter, "test", "10")
		require.NoError(t, err)
		memMetric := model.NewMetricCounter("test", 1)
		want := memMetric.Clone()
		*want.Delta += *mr.Delta
		mockRepository.EXPECT().Find(gomock.Eq(mr)).Return(memMetric, nil)
		mockRepository.EXPECT().Create(gomock.Any()).MaxTimes(0)
		mockRepository.EXPECT().Update(gomock.Eq(&model.MetricRequest{Metric: want})).Return(want, nil)
		service := NewService(mockRepository)
		got, err := service.Update(mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})
}
