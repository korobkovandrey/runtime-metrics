package pbservice

import (
	"context"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	pb "github.com/korobkovandrey/runtime-metrics/internal/proto"
	"github.com/korobkovandrey/runtime-metrics/internal/server/factory"
	"github.com/korobkovandrey/runtime-metrics/internal/server/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type updater interface {
	Update(context.Context, *model.MetricRequest) (*model.Metric, error)
}
type batchUpdater interface {
	UpdateBatch(context.Context, []*model.MetricRequest) ([]*model.Metric, error)
}

type metricsService struct {
	updater
	batchUpdater
}

type MetricsService struct {
	pb.UnimplementedMetricsServiceServer
	s *metricsService
}

func NewMetricsService(r factory.Repository) *MetricsService {
	return &MetricsService{s: &metricsService{
		updater:      service.NewUpdater(r),
		batchUpdater: service.NewBatchUpdater(r),
	}}
}

// Update method
func (ms *MetricsService) Update(ctx context.Context, req *pb.Metric) (*pb.Response, error) {
	mr, err := pb.MetricToModelMetricRequest(req)
	if err == nil {
		err = mr.RequiredValue()
	}
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if _, err = ms.s.Update(ctx, mr); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Response{}, nil
}

// Updates method
func (ms *MetricsService) Updates(ctx context.Context, req *pb.Metrics) (*pb.Response, error) {
	mrs, err := pb.MetricsToModelsMetricRequest(req.Metrics)
	if err == nil {
		err = model.ValidateMetricsRequest(mrs)
	}
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if _, err = ms.s.UpdateBatch(ctx, mrs); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Response{}, nil
}
