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

func TestFinder_Find(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("valid", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "1")
		require.NoError(t, err)
		want := mr.Clone()
		r := mocks.NewMockFinderRepository(ctrl)
		r.EXPECT().Find(gomock.Any(), gomock.Eq(mr)).Return(want, nil)
		s := NewFinder(r)
		got, err := s.Find(context.TODO(), mr)
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("error", func(t *testing.T) {
		mr, err := model.NewMetricRequest(model.TypeGauge, "test", "1")
		require.NoError(t, err)
		r := mocks.NewMockFinderRepository(ctrl)
		r.EXPECT().Find(gomock.Any(), gomock.Eq(mr)).Return(nil, model.ErrMetricNotFound)
		s := NewFinder(r)
		got, err := s.Find(context.TODO(), mr)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, model.ErrMetricNotFound)
	})
}

func TestFinder_FindAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("valid", func(t *testing.T) {
		want := []*model.Metric{
			model.NewMetricGauge("test1", 1),
			model.NewMetricCounter("test2", 1),
		}
		r := mocks.NewMockFinderRepository(ctrl)
		r.EXPECT().FindAll(gomock.Any()).Return(want, nil)
		s := NewFinder(r)
		got, err := s.FindAll(context.TODO())
		assert.NoError(t, err)
		assert.ElementsMatch(t, want, got)
	})

	t.Run("error", func(t *testing.T) {
		r := mocks.NewMockFinderRepository(ctrl)
		r.EXPECT().FindAll(gomock.Any()).
			Return([]*model.Metric{}, errors.New("error"))
		s := NewFinder(r)
		got, err := s.FindAll(context.TODO())
		assert.Nil(t, got)
		assert.Error(t, err)
	})
}
