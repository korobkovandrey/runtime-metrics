package model

import (
	"testing"

	"github.com/korobkovandrey/runtime-metrics/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricToModelMetricRequest(t *testing.T) {
	tests := []struct {
		input   *proto.Metric
		want    *MetricRequest
		name    string
		errMsg  string
		wantErr bool
	}{
		{
			name: "valid counter",
			input: &proto.Metric{
				Id:    "counter1",
				Type:  proto.MetricType_COUNTER,
				Delta: 65,
			},
			want: NewMetricCounter("counter1", 65).ToRequest(),
		},
		{
			name: "valid gauge",
			input: &proto.Metric{
				Id:    "gauge1",
				Type:  proto.MetricType_GAUGE,
				Value: 12.34,
			},
			want: NewMetricGauge("gauge1", 12.34).ToRequest(),
		},
		{
			name: "invalid type",
			input: &proto.Metric{
				Id:   "test",
				Type: proto.MetricType_UNKNOWN,
			},
			wantErr: true,
			errMsg:  ErrTypeIsNotValid.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MetricToModelMetricRequest(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetricsToModelsMetricRequest(t *testing.T) {
	tests := []struct {
		name    string
		errMsg  string
		input   []*proto.Metric
		want    []*MetricRequest
		wantErr bool
	}{
		{
			name: "valid metrics",
			input: []*proto.Metric{
				{
					Id:    "counter1",
					Type:  proto.MetricType_COUNTER,
					Delta: 65,
				},
				{
					Id:    "gauge1",
					Type:  proto.MetricType_GAUGE,
					Value: 12.34,
				},
			},
			want: []*MetricRequest{
				NewMetricCounter("counter1", 65).ToRequest(),
				NewMetricGauge("gauge1", 12.34).ToRequest(),
			},
		},
		{
			name:  "empty metrics",
			input: []*proto.Metric{},
			want:  []*MetricRequest{},
		},
		{
			name: "invalid metric type",
			input: []*proto.Metric{
				{
					Id:    "counter1",
					Type:  proto.MetricType_COUNTER,
					Delta: 65,
				},
				{
					Id:   "test",
					Type: proto.MetricType_UNKNOWN,
				},
			},
			wantErr: true,
			errMsg:  ErrTypeIsNotValid.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MetricsToModelsMetricRequest(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestModelMetricToMetric(t *testing.T) {
	tests := []struct {
		input *Metric
		want  *proto.Metric
		name  string
	}{
		{
			name:  "valid counter",
			input: NewMetricCounter("counter1", 65),
			want: &proto.Metric{
				Id:    "counter1",
				Type:  proto.MetricType_COUNTER,
				Delta: 65,
			},
		},
		{
			name:  "valid gauge",
			input: NewMetricGauge("gauge1", 12.34),
			want: &proto.Metric{
				Id:    "gauge1",
				Type:  proto.MetricType_GAUGE,
				Value: 12.34,
			},
		},
		{
			name: "nil counter delta",
			input: &Metric{
				ID:    "counter1",
				MType: TypeCounter,
			},
			want: &proto.Metric{
				Id:   "counter1",
				Type: proto.MetricType_COUNTER,
			},
		},
		{
			name: "nil gauge value",
			input: &Metric{
				ID:    "gauge1",
				MType: TypeGauge,
			},
			want: &proto.Metric{
				Id:   "gauge1",
				Type: proto.MetricType_GAUGE,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ModelMetricToMetric(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestModelMetricsToMetrics(t *testing.T) {
	tests := []struct {
		name  string
		input []*Metric
		want  []*proto.Metric
	}{
		{
			name: "valid metrics",
			input: []*Metric{
				NewMetricCounter("counter1", 65),
				NewMetricGauge("gauge1", 12.34),
			},
			want: []*proto.Metric{
				{
					Id:    "counter1",
					Type:  proto.MetricType_COUNTER,
					Delta: 65,
				},
				{
					Id:    "gauge1",
					Type:  proto.MetricType_GAUGE,
					Value: 12.34,
				},
			},
		},
		{
			name:  "empty metrics",
			input: []*Metric{},
			want:  []*proto.Metric{},
		},
		{
			name: "nil values",
			input: []*Metric{
				{
					ID:    "counter1",
					MType: TypeCounter,
				},
				{
					ID:    "gauge1",
					MType: TypeGauge,
				},
			},
			want: []*proto.Metric{
				{
					Id:   "counter1",
					Type: proto.MetricType_COUNTER,
				},
				{
					Id:   "gauge1",
					Type: proto.MetricType_GAUGE,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ModelMetricsToMetrics(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
