package proto

import (
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricToModelMetricRequest(t *testing.T) {
	tests := []struct {
		input   *Metric
		want    *model.MetricRequest
		name    string
		errMsg  string
		wantErr bool
	}{
		{
			name: "valid counter",
			input: &Metric{
				Id:    "counter1",
				Type:  MetricType_COUNTER,
				Delta: 65,
			},
			want: model.NewMetricCounter("counter1", 65).ToRequest(),
		},
		{
			name: "valid gauge",
			input: &Metric{
				Id:    "gauge1",
				Type:  MetricType_GAUGE,
				Value: 12.34,
			},
			want: model.NewMetricGauge("gauge1", 12.34).ToRequest(),
		},
		{
			name: "invalid type",
			input: &Metric{
				Id:   "test",
				Type: MetricType_UNKNOWN,
			},
			wantErr: true,
			errMsg:  model.ErrTypeIsNotValid.Error(),
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
		input   []*Metric
		want    []*model.MetricRequest
		wantErr bool
	}{
		{
			name: "valid metrics",
			input: []*Metric{
				{
					Id:    "counter1",
					Type:  MetricType_COUNTER,
					Delta: 65,
				},
				{
					Id:    "gauge1",
					Type:  MetricType_GAUGE,
					Value: 12.34,
				},
			},
			want: []*model.MetricRequest{
				model.NewMetricCounter("counter1", 65).ToRequest(),
				model.NewMetricGauge("gauge1", 12.34).ToRequest(),
			},
		},
		{
			name:  "empty metrics",
			input: []*Metric{},
			want:  []*model.MetricRequest{},
		},
		{
			name: "invalid metric type",
			input: []*Metric{
				{
					Id:    "counter1",
					Type:  MetricType_COUNTER,
					Delta: 65,
				},
				{
					Id:   "test",
					Type: MetricType_UNKNOWN,
				},
			},
			wantErr: true,
			errMsg:  model.ErrTypeIsNotValid.Error(),
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
		input *model.Metric
		want  *Metric
		name  string
	}{
		{
			name:  "valid counter",
			input: model.NewMetricCounter("counter1", 65),
			want: &Metric{
				Id:    "counter1",
				Type:  MetricType_COUNTER,
				Delta: 65,
			},
		},
		{
			name:  "valid gauge",
			input: model.NewMetricGauge("gauge1", 12.34),
			want: &Metric{
				Id:    "gauge1",
				Type:  MetricType_GAUGE,
				Value: 12.34,
			},
		},
		{
			name: "nil counter delta",
			input: &model.Metric{
				ID:    "counter1",
				MType: model.TypeCounter,
			},
			want: &Metric{
				Id:   "counter1",
				Type: MetricType_COUNTER,
			},
		},
		{
			name: "nil gauge value",
			input: &model.Metric{
				ID:    "gauge1",
				MType: model.TypeGauge,
			},
			want: &Metric{
				Id:   "gauge1",
				Type: MetricType_GAUGE,
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
		input []*model.Metric
		want  []*Metric
	}{
		{
			name: "valid metrics",
			input: []*model.Metric{
				model.NewMetricCounter("counter1", 65),
				model.NewMetricGauge("gauge1", 12.34),
			},
			want: []*Metric{
				{
					Id:    "counter1",
					Type:  MetricType_COUNTER,
					Delta: 65,
				},
				{
					Id:    "gauge1",
					Type:  MetricType_GAUGE,
					Value: 12.34,
				},
			},
		},
		{
			name:  "empty metrics",
			input: []*model.Metric{},
			want:  []*Metric{},
		},
		{
			name: "nil values",
			input: []*model.Metric{
				{
					ID:    "counter1",
					MType: model.TypeCounter,
				},
				{
					ID:    "gauge1",
					MType: model.TypeGauge,
				},
			},
			want: []*Metric{
				{
					Id:   "counter1",
					Type: MetricType_COUNTER,
				},
				{
					Id:   "gauge1",
					Type: MetricType_GAUGE,
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
