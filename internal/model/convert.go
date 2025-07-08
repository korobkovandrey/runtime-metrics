package model

import (
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/pkg/proto"
)

func MetricToModelMetricRequest(x *proto.Metric) (*MetricRequest, error) {
	switch x.GetType() {
	case proto.MetricType_COUNTER:
		return NewMetricCounter(x.GetId(), x.GetDelta()).ToRequest(), nil
	case proto.MetricType_GAUGE:
		return NewMetricGauge(x.GetId(), x.GetValue()).ToRequest(), nil
	}
	return nil, fmt.Errorf("%w: %s", ErrTypeIsNotValid, x.GetType().String())
}

func MetricsToModelsMetricRequest(xs []*proto.Metric) ([]*MetricRequest, error) {
	mrs := make([]*MetricRequest, len(xs))
	for i := range xs {
		mr, err := MetricToModelMetricRequest(xs[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert metric: %w", err)
		}
		mrs[i] = mr
	}
	return mrs, nil
}

func ModelMetricToMetric(m *Metric) *proto.Metric {
	mm := &proto.Metric{Id: m.ID}
	switch m.MType {
	case TypeCounter:
		mm.Type = proto.MetricType_COUNTER
		if m.Delta != nil {
			mm.Delta = *m.Delta
		}
	case TypeGauge:
		mm.Type = proto.MetricType_GAUGE
		if m.Value != nil {
			mm.Value = *m.Value
		}
	}
	return mm
}

func ModelMetricsToMetrics(ms []*Metric) []*proto.Metric {
	xs := make([]*proto.Metric, len(ms))
	for i := range ms {
		xs[i] = ModelMetricToMetric(ms[i])
	}
	return xs
}
