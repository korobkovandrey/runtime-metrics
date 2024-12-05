package utils

import (
	"github.com/stretchr/testify/assert"

	"reflect"
	"strings"
	"testing"
)

func Test_convertReflectValueToString(t *testing.T) {
	type args struct {
		value reflect.Value
	}
	v1 := int64(10)
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "positive int64",
			args: args{
				value: reflect.ValueOf(int64(10)),
			},
			want:    "10",
			wantErr: nil,
		},
		{
			name: "negative int64",
			args: args{
				value: reflect.ValueOf(int64(-10)),
			},
			want:    "-10",
			wantErr: nil,
		},
		{
			name: "uint64",
			args: args{
				value: reflect.ValueOf(uint64(10)),
			},
			want:    "10",
			wantErr: nil,
		},
		{
			name: "float64",
			args: args{
				value: reflect.ValueOf(float64(10.123345)),
			},
			want:    "10.123345",
			wantErr: nil,
		},
		{
			name: "string float64",
			args: args{
				value: reflect.ValueOf("10.123345"),
			},
			want:    "10.123345",
			wantErr: nil,
		},
		{
			name: "fail numeric string",
			args: args{
				value: reflect.ValueOf("10.123.345"),
			},
			want:    "",
			wantErr: ErrNumberIsNotNumber,
		},
		{
			name: "nil",
			args: args{
				value: reflect.ValueOf(nil),
			},
			want:    "",
			wantErr: ErrNumberIsNotNumber,
		},
		{
			name: "bool",
			args: args{
				value: reflect.ValueOf(true),
			},
			want:    "1",
			wantErr: nil,
		},
		{
			name: "pointer int64",
			args: args{
				value: reflect.ValueOf(&v1),
			},
			want:    "10",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertReflectValueToString(tt.args.value)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			if tt.wantErr == nil {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetRuntimeMetrics(t *testing.T) {
	metricsOk := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle",
		"HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs",
		"StackInuse", "StackSys", "Sys", "TotalAlloc"}
	notNumberMetrics := []string{"PauseEnd", "PauseNs", "BySize"}
	notFoundMetrics := []string{"NotFoundMetric"}
	allMetrics := metricsOk
	allMetrics = append(allMetrics, notNumberMetrics...)
	allMetrics = append(allMetrics, notFoundMetrics...)
	gotResult, gotErrNotNumber, gotErrNotFound := GetRuntimeMetrics(allMetrics...)
	for _, m := range metricsOk {
		assert.Contains(t, gotResult, m)
	}
	for m := range gotResult {
		assert.Contains(t, metricsOk, m)
	}
	assert.ErrorIs(t, gotErrNotNumber, ErrNumberIsNotNumber)
	assert.ErrorIs(t, gotErrNotFound, ErrFieldNotFound)
	assert.Equal(t, gotErrNotNumber.Error(), `"`+strings.Join(notNumberMetrics, `", "`)+`": `+ErrNumberIsNotNumber.Error())
	assert.Equal(t, gotErrNotFound.Error(), `"`+strings.Join(notFoundMetrics, `", "`)+`": `+ErrFieldNotFound.Error())
}
