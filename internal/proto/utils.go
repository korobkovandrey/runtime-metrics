package proto

import (
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

func MetricToModelMetricRequest(x *Metric) (*model.MetricRequest, error) {
	switch x.GetType() {
	case MetricType_COUNTER:
		return model.NewMetricCounter(x.GetId(), x.GetDelta()).ToRequest(), nil
	case MetricType_GAUGE:
		return model.NewMetricGauge(x.GetId(), x.GetValue()).ToRequest(), nil
	}
	return nil, fmt.Errorf("%w: %s", model.ErrTypeIsNotValid, x.GetType().String())
}

func MetricsToModelsMetricRequest(xs []*Metric) ([]*model.MetricRequest, error) {
	mrs := make([]*model.MetricRequest, len(xs))
	for i := range xs {
		mr, err := MetricToModelMetricRequest(xs[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert metric: %w", err)
		}
		mrs[i] = mr
	}
	return mrs, nil
}

func ModelMetricToMetric(m *model.Metric) *Metric {
	mm := &Metric{Id: m.ID}
	switch m.MType {
	case model.TypeCounter:
		mm.Type = MetricType_COUNTER
		if m.Delta != nil {
			mm.Delta = *m.Delta
		}
	case model.TypeGauge:
		mm.Type = MetricType_GAUGE
		if m.Value != nil {
			mm.Value = *m.Value
		}
	}
	return mm
}

func ModelMetricsToMetrics(ms []*model.Metric) []*Metric {
	xs := make([]*Metric, len(ms))
	for i := range ms {
		xs[i] = ModelMetricToMetric(ms[i])
	}
	return xs
}
