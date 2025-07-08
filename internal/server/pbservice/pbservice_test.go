package pbservice

import (
	"context"
	"errors"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/handlers/mocks"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	pb "github.com/korobkovandrey/runtime-metrics/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewMetricsService(t *testing.T) {
	require.NotPanics(t, func() {
		NewMetricsService(repository.NewMemStorage())
	})
}

func TestMetricsService_Update(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("ValidCounterRequest", func(t *testing.T) {
		mockUpdater := mocks.NewMockUpdater(ctrl)
		ms := &MetricsService{
			s: &metricsService{
				updater: mockUpdater,
			},
		}
		req := &pb.Metric{
			Id:    "test-counter",
			Type:  pb.MetricType_COUNTER,
			Delta: 42,
		}
		mr := model.NewMetricCounter("test-counter", 42).ToRequest()
		expectedMetric := model.NewMetricCounter("test-counter", 42)
		mockUpdater.EXPECT().
			Update(ctx, mr).
			Return(expectedMetric, nil).
			Times(1)

		resp, err := ms.Update(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, &pb.Response{}, resp)
	})

	t.Run("ValidGaugeRequest", func(t *testing.T) {
		mockUpdater := mocks.NewMockUpdater(ctrl)
		ms := &MetricsService{
			s: &metricsService{
				updater: mockUpdater,
			},
		}
		req := &pb.Metric{
			Id:    "test-gauge",
			Type:  pb.MetricType_GAUGE,
			Value: 3.14,
		}
		mr := model.NewMetricGauge("test-gauge", 3.14).ToRequest()
		expectedMetric := model.NewMetricGauge("test-gauge", 3.14)
		mockUpdater.EXPECT().
			Update(ctx, mr).
			Return(expectedMetric, nil).
			Times(1)

		resp, err := ms.Update(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, &pb.Response{}, resp)
	})

	t.Run("InvalidProtoMetric", func(t *testing.T) {
		mockUpdater := mocks.NewMockUpdater(ctrl)
		ms := &MetricsService{
			s: &metricsService{
				updater: mockUpdater,
			},
		}
		req := &pb.Metric{
			Id:    "test-metric",
			Type:  pb.MetricType(999),
			Value: 1,
		}
		resp, err := ms.Update(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), model.ErrTypeIsNotValid.Error())
	})

	t.Run("UpdaterError", func(t *testing.T) {
		mockUpdater := mocks.NewMockUpdater(ctrl)
		ms := &MetricsService{
			s: &metricsService{
				updater: mockUpdater,
			},
		}
		req := &pb.Metric{
			Id:    "test-counter",
			Type:  pb.MetricType_COUNTER,
			Delta: 42,
		}
		mr := model.NewMetricCounter("test-counter", 42).ToRequest()
		mockUpdater.EXPECT().
			Update(ctx, mr).
			Return(nil, errors.New("update failed")).
			Times(1)

		resp, err := ms.Update(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Contains(t, err.Error(), "update failed")
	})
}

func TestMetricsService_Updates(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("ValidBatchRequest", func(t *testing.T) {
		mockBatchUpdater := mocks.NewMockBatchUpdater(ctrl)
		ms := &MetricsService{
			s: &metricsService{
				batchUpdater: mockBatchUpdater,
			},
		}
		req := &pb.Metrics{
			Metrics: []*pb.Metric{
				{Id: "test-counter", Type: pb.MetricType_COUNTER, Delta: 42},
				{Id: "test-gauge", Type: pb.MetricType_GAUGE, Value: 3.14},
			},
		}
		mrs := []*model.MetricRequest{
			model.NewMetricCounter("test-counter", 42).ToRequest(),
			model.NewMetricGauge("test-gauge", 3.14).ToRequest(),
		}
		expectedMetrics := []*model.Metric{
			model.NewMetricCounter("test-counter", 42),
			model.NewMetricGauge("test-gauge", 3.14),
		}
		mockBatchUpdater.EXPECT().
			UpdateBatch(ctx, mrs).
			Return(expectedMetrics, nil).
			Times(1)

		resp, err := ms.Updates(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, &pb.Response{}, resp)
	})

	t.Run("InvalidProtoMetrics", func(t *testing.T) {
		mockBatchUpdater := mocks.NewMockBatchUpdater(ctrl)
		ms := &MetricsService{
			s: &metricsService{
				batchUpdater: mockBatchUpdater,
			},
		}
		req := &pb.Metrics{
			Metrics: []*pb.Metric{
				{Id: "test-metric", Type: pb.MetricType(999)},
			},
		}
		resp, err := ms.Updates(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), model.ErrTypeIsNotValid.Error())
	})

	t.Run("BatchUpdaterError", func(t *testing.T) {
		mockBatchUpdater := mocks.NewMockBatchUpdater(ctrl)
		ms := &MetricsService{
			s: &metricsService{
				batchUpdater: mockBatchUpdater,
			},
		}
		req := &pb.Metrics{
			Metrics: []*pb.Metric{
				{Id: "test-counter", Type: pb.MetricType_COUNTER, Delta: 42},
				{Id: "test-gauge", Type: pb.MetricType_GAUGE, Value: 3.14},
			},
		}
		mrs := []*model.MetricRequest{
			model.NewMetricCounter("test-counter", 42).ToRequest(),
			model.NewMetricGauge("test-gauge", 3.14).ToRequest(),
		}
		mockBatchUpdater.EXPECT().
			UpdateBatch(ctx, mrs).
			Return(nil, errors.New("batch update failed")).
			Times(1)

		resp, err := ms.Updates(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Contains(t, err.Error(), "batch update failed")
	})
}
