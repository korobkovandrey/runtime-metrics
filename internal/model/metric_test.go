package model

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pointer[T any](v T) *T {
	return &v
}

func TestMetric_AnyValue(t *testing.T) {
	type args struct {
		Metric *Metric
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "gauge",
			args: args{
				Metric: NewMetricGauge("test", 76),
			},
			want: 76.0,
		},
		{
			name: "counter",
			args: args{
				Metric: NewMetricCounter("test", 15),
			},
			want: int64(15),
		},
		{
			name: "nil",
			args: args{
				Metric: &Metric{
					MType: "test",
					ID:    "test",
				},
			},
			want: nil,
		},
		{
			name: "nil delta",
			args: args{
				Metric: &Metric{
					MType: TypeCounter,
					ID:    "test",
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.args.Metric.AnyValue())
		})
	}
}

func TestMetric_Clone(t *testing.T) {
	type args struct {
		Metric *Metric
	}
	tests := []struct {
		name string
		args args
		want *Metric
	}{
		{
			name: "gauge",
			args: args{
				Metric: NewMetricGauge("test", 64),
			},
			want: NewMetricGauge("test", 64),
		},
		{
			name: "counter",
			args: args{
				Metric: NewMetricCounter("test", 33),
			},
			want: NewMetricCounter("test", 33),
		},
		{
			name: "nil",
			args: args{
				Metric: &Metric{
					MType: "test",
					ID:    "test",
				},
			},
			want: &Metric{
				MType: "test",
				ID:    "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Metric.Clone()
			assert.Equal(t, tt.want, got)
			assert.NotSame(t, tt.want, got)
			if tt.want.Value != nil {
				assert.NotNil(t, got.Value)
				assert.NotSame(t, tt.want.Value, got.Value)
			}
			if tt.want.Delta != nil {
				assert.NotNil(t, got.Delta)
				assert.NotSame(t, tt.want.Delta, got.Delta)
			}
		})
	}
}

func TestNewMetricCounter(t *testing.T) {
	type args struct {
		id    string
		delta int64
	}
	tests := []struct {
		name string
		args args
		want *Metric
	}{
		{
			name: "positive delta",
			args: args{
				id:    "test",
				delta: 14,
			},
			want: &Metric{
				MType: TypeCounter,
				ID:    "test",
				Delta: pointer(int64(14)),
			},
		},
		{
			name: "negative delta",
			args: args{
				id:    "test",
				delta: -12,
			},
			want: &Metric{
				MType: TypeCounter,
				ID:    "test",
				Delta: pointer(int64(-12)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetricCounter(tt.args.id, tt.args.delta)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMetricGauge(t *testing.T) {
	type args struct {
		id    string
		value float64
	}
	tests := []struct {
		name string
		args args
		want *Metric
	}{
		{
			name: "positive value",
			args: args{
				id:    "test",
				value: 12.33,
			},
			want: &Metric{
				MType: TypeGauge,
				ID:    "test",
				Value: pointer(12.33),
			},
		},
		{
			name: "negative value",
			args: args{
				id:    "test",
				value: -12.33,
			},
			want: &Metric{
				MType: TypeGauge,
				ID:    "test",
				Value: pointer(-12.33),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetricGauge(tt.args.id, tt.args.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMetricRequest(t *testing.T) {
	type args struct {
		t     string
		name  string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    *MetricRequest
		wantErr error
	}{
		{
			name: "counter",
			args: args{
				t:     TypeCounter,
				name:  "test",
				value: "42",
			},
			want: &MetricRequest{NewMetricCounter("test", 42)},
		},
		{
			name: "gauge",
			args: args{
				t:     TypeGauge,
				name:  "test",
				value: "12.34",
			},
			want: &MetricRequest{NewMetricGauge("test", 12.34)},
		},
		{
			name: "unknown type",
			args: args{
				t:     "unknown",
				name:  "test",
				value: "42",
			},
			wantErr: ErrTypeIsNotValid,
		},
		{
			name: "empty name",
			args: args{
				t:     TypeCounter,
				name:  "",
				value: "42",
			},
			want: &MetricRequest{NewMetricCounter("", 42)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMetricRequest(tt.args.t, tt.args.name, tt.args.value)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnmarshalMetricRequestFromReader(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *MetricRequest
		wantErr error
	}{
		{
			name: "counter from reader",
			args: args{
				r: strings.NewReader(`{"type":"counter","id":"test","delta":65}`),
			},
			want: &MetricRequest{NewMetricCounter("test", 65)},
		},
		{
			name: "gauge from reader",
			args: args{
				r: strings.NewReader(`{"type":"gauge","id":"test","value":12.34}`),
			},
			want: &MetricRequest{NewMetricGauge("test", 12.34)},
		},
		{
			name: "error",
			args: args{
				r: strings.NewReader(`{}`),
			},
			wantErr: ErrMetricNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalMetricRequestFromReader(tt.args.r)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
