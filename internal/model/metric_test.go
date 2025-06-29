package model

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/pashagolub/pgxmock/v4"
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
		want any
		args args
		name string
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
		args args
		want *Metric
		name string
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
		want *Metric
		name string
		args args
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
		want *Metric
		name string
		args args
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
		wantErr error
		want    *MetricRequest
		args    args
		name    string
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
			name: "counter error number",
			args: args{
				t:     TypeCounter,
				name:  "test",
				value: "invalid",
			},
			wantErr: ErrValueIsNotValid,
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
			name: "gauge error number",
			args: args{
				t:     TypeGauge,
				name:  "test",
				value: "invalid",
			},
			wantErr: ErrValueIsNotValid,
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
	anyError := errors.New("any error")
	errorReader := errors.New("error reader")
	tests := []struct {
		args    args
		wantErr error
		want    *MetricRequest
		name    string
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
		{
			name: "empty body",
			args: args{
				r: strings.NewReader(""),
			},
			wantErr: anyError,
		},
		{
			name: "error unmarshal",
			args: args{
				r: strings.NewReader(`{`),
			},
			wantErr: anyError,
		},
		{
			name: "error reader",
			args: args{
				r: iotest.ErrReader(errorReader),
			},
			wantErr: errorReader,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalMetricRequestFromReader(tt.args.r)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				if !errors.Is(tt.wantErr, anyError) {
					assert.ErrorIs(t, err, tt.wantErr)
				}
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetricRequest_RequiredValue(t *testing.T) {
	tests := []struct {
		wantErr error
		metric  *MetricRequest
		name    string
	}{
		{
			name:    "gauge",
			metric:  &MetricRequest{NewMetricGauge("test", 12.34)},
			wantErr: nil,
		},
		{
			name:    "gauge error",
			metric:  &MetricRequest{&Metric{MType: TypeGauge, ID: "test"}},
			wantErr: ErrValueIsNotValid,
		},
		{
			name:    "counter",
			metric:  &MetricRequest{NewMetricCounter("test", 12)},
			wantErr: nil,
		},
		{
			name:    "counter error",
			metric:  &MetricRequest{&Metric{MType: TypeCounter, ID: "test"}},
			wantErr: ErrValueIsNotValid,
		},
		{
			name:    "type error",
			metric:  &MetricRequest{&Metric{MType: "invalid", ID: "test"}},
			wantErr: ErrTypeIsNotValid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.RequiredValue()
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestMetricRequest_ValidateType(t *testing.T) {
	tests := []struct {
		wantErr error
		metric  *MetricRequest
		name    string
	}{
		{
			name:    "gauge",
			metric:  &MetricRequest{NewMetricGauge("test", 12.34)},
			wantErr: nil,
		},
		{
			name:    "counter",
			metric:  &MetricRequest{NewMetricCounter("test", 12)},
			wantErr: nil,
		},
		{
			name:    "type error",
			metric:  &MetricRequest{&Metric{MType: "invalid", ID: "test"}},
			wantErr: ErrTypeIsNotValid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.ValidateType()
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestUnmarshalMetricsRequestFromReader(t *testing.T) {
	type args struct {
		r io.Reader
	}
	anyError := errors.New("any error")
	errorReader := errors.New("error reader")
	tests := []struct {
		args    args
		wantErr error
		name    string
		want    []*MetricRequest
	}{
		{
			name: "valid metrics list",
			args: args{
				r: strings.NewReader(`[
					{"type":"counter","id":"counter1","delta":65},
					{"type":"gauge","id":"gauge1","value":12.34}
				]`),
			},
			want: []*MetricRequest{
				{Metric: NewMetricCounter("counter1", 65)},
				{Metric: NewMetricGauge("gauge1", 12.34)},
			},
		},
		{
			name: "empty body",
			args: args{
				r: strings.NewReader(``),
			},
			wantErr: anyError,
		},
		{
			name: "invalid JSON",
			args: args{
				r: strings.NewReader(`[{invalid}]`),
			},
			wantErr: anyError,
		},
		{
			name: "empty metrics list",
			args: args{
				r: strings.NewReader(`[]`),
			},
			wantErr: anyError,
		},
		{
			name: "error reader",
			args: args{
				r: iotest.ErrReader(errorReader),
			},
			wantErr: errorReader,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalMetricsRequestFromReader(tt.args.r)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.Error(t, err)
				if !errors.Is(tt.wantErr, anyError) {
					assert.ErrorIs(t, err, tt.wantErr)
				}
			}
		})
	}
}

func TestValidateMetricsRequest(t *testing.T) {
	tests := []struct {
		wantErr error
		name    string
		metrics []*MetricRequest
	}{
		{
			name: "valid metrics",
			metrics: []*MetricRequest{
				{Metric: NewMetricCounter("counter1", 65)},
				{Metric: NewMetricGauge("gauge1", 12.34)},
			},
			wantErr: nil,
		},
		{
			name: "invalid gauge value",
			metrics: []*MetricRequest{
				{Metric: NewMetricCounter("counter1", 65)},
				{Metric: &Metric{MType: TypeGauge, ID: "gauge1"}},
			},
			wantErr: ErrValueIsNotValid,
		},
		{
			name: "invalid counter value",
			metrics: []*MetricRequest{
				{Metric: NewMetricGauge("gauge1", 12.34)},
				{Metric: &Metric{MType: TypeCounter, ID: "counter1"}},
			},
			wantErr: ErrValueIsNotValid,
		},
		{
			name: "invalid type",
			metrics: []*MetricRequest{
				{Metric: NewMetricGauge("gauge1", 12.34)},
				{Metric: &Metric{MType: "invalid", ID: "test"}},
			},
			wantErr: ErrTypeIsNotValid,
		},
		{
			name:    "empty metrics list",
			metrics: []*MetricRequest{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetricsRequest(tt.metrics)
			if tt.wantErr == nil {
				assert.NoError(t, err, "Expected no error")
			} else {
				require.Error(t, err, "Expected error")
				assert.ErrorIs(t, err, tt.wantErr, "Unexpected error")
			}
		})
	}
}

func TestMetric_ScanRow(t *testing.T) {
	tests := []struct {
		setupMock   func(mock pgxmock.PgxPoolIface)
		metric      *Metric
		wantMetric  *Metric
		name        string
		errContains string
		wantErr     bool
	}{
		{
			name: "gauge",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"mtype", "id", "value", "delta"}).
					AddRow(TypeGauge, "gauge1", pointer(12.34), nil)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			metric: &Metric{},
			wantMetric: &Metric{
				MType: TypeGauge,
				ID:    "gauge1",
				Value: pointer(12.34),
				Delta: nil,
			},
			wantErr: false,
		},
		{
			name: "counter",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"mtype", "id", "value", "delta"}).
					AddRow(TypeCounter, "counter1", nil, pointer(int64(65)))
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			metric: &Metric{},
			wantMetric: &Metric{
				MType: TypeCounter,
				ID:    "counter1",
				Value: nil,
				Delta: pointer(int64(65)),
			},
			wantErr: false,
		},
		{
			name: "scan error",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT").WillReturnError(errors.New("database error"))
			},
			metric:      &Metric{},
			wantMetric:  &Metric{},
			wantErr:     true,
			errContains: "database error",
		},
		{
			name: "invalid type",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"mtype", "id", "value", "delta"}).
					AddRow("invalid", "test", pointer(12.34), nil)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			metric: &Metric{},
			wantMetric: &Metric{
				MType: "invalid",
				ID:    "test",
				Value: pointer(12.34),
				Delta: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.setupMock(mock)
			row, err := mock.Query(t.Context(), "SELECT")
			if tt.name == "scan error" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			} else {
				require.NoError(t, err)
			}
			defer row.Close()

			if row.Next() {
				err = tt.metric.ScanRow(row)
			} else if !tt.wantErr {
				t.Fatal("No rows returned by mock query")
			}

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMetric, tt.metric)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
