package service

import (
	"context"
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
	assert.Equal(t, &Service{r: mockRepository}, NewService(mockRepository))
}

func TestService_Find(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepository := mocks.NewMockRepository(ctrl)

	t.Run("valid", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "1")
		require.NoError(t, err)
		want := mr.Clone()
		ctx := context.TODO()
		mockRepository.EXPECT().Find(gomock.Eq(ctx), gomock.Eq(mr)).Return(want, nil)
		service := NewService(mockRepository)
		got, err := service.Find(ctx, mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("error", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "1")
		require.NoError(t, err)
		ctx := context.TODO()
		mockRepository.EXPECT().Find(gomock.Eq(ctx), gomock.Eq(mr)).Return(nil, model.ErrMetricNotFound)
		service := NewService(mockRepository)
		got, err := service.Find(ctx, mr)
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
		ctx := context.TODO()
		mockRepository.EXPECT().FindAll(gomock.Eq(ctx)).Return(want, nil)
		service := NewService(mockRepository)
		got, err := service.FindAll(ctx)
		assert.NoError(t, err)
		assert.ElementsMatch(t, want, got)
	})

	t.Run("error", func(t *testing.T) {
		ctx := context.TODO()
		mockRepository.EXPECT().FindAll(gomock.Eq(ctx)).
			Return([]*model.Metric{}, errors.New("error"))
		service := NewService(mockRepository)
		got, err := service.FindAll(ctx)
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
		ctx := context.TODO()
		mockRepository.EXPECT().Find(gomock.Eq(ctx), gomock.Eq(mr)).Return(nil, model.ErrMetricNotFound)
		mockRepository.EXPECT().Create(gomock.Eq(ctx), gomock.Eq(mr)).Return(want, nil)
		mockRepository.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
		service := NewService(mockRepository)
		got, err := service.Update(ctx, mr)
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
		mockRepository.EXPECT().Find(gomock.Eq(ctx), gomock.Eq(mr)).Return(memMetric, nil)
		mockRepository.EXPECT().Create(gomock.Any(), gomock.Any()).MaxTimes(0)
		mockRepository.EXPECT().Update(gomock.Eq(ctx), gomock.Eq(&model.MetricRequest{Metric: want})).Return(want, nil)
		service := NewService(mockRepository)
		got, err := service.Update(ctx, mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})
}

func TestService_UpdateBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepository := mocks.NewMockRepository(ctrl)

	t.Run("update batch", func(t *testing.T) {
		mrs := []*model.MetricRequest{
			{Metric: model.NewMetricGauge("testNotExist", 12.34)},
			{Metric: model.NewMetricCounter("testNotExist", 1)},
			{Metric: model.NewMetricGauge("testNotExist", 22.55)},
			{Metric: model.NewMetricCounter("testNotExist", 2)},
			{Metric: model.NewMetricGauge("testExist", 55.44)},
			{Metric: model.NewMetricCounter("testExist", 5)},
			{Metric: model.NewMetricGauge("testExist", 66.7)},
			{Metric: model.NewMetricCounter("testExist", 8)},
		}
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
		ctx := context.TODO()
		mockRepository.EXPECT().FindBatch(gomock.Eq(ctx), gomock.Eq([]*model.MetricRequest{
			{Metric: model.NewMetricCounter("testNotExist", 3)},
			{Metric: model.NewMetricCounter("testExist", 13)},
		})).Return([]*model.Metric{model.NewMetricCounter("testExist", 3)}, nil)
		mockRepository.EXPECT().CreateOrUpdateBatch(gomock.Eq(ctx), gomock.Eq(mrsReq)).Return(want, nil)
		service := NewService(mockRepository)
		got, err := service.UpdateBatch(ctx, mrs)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
