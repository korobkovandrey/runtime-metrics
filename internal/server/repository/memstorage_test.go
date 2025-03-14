package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemStorage(t *testing.T) {
	assert.Equal(t, newMemStorageWithDataAndIndex([]*model.Metric{}, map[string]map[string]int{}), NewMemStorage())
}

func TestMemStorage_Create(t *testing.T) {
	type fields struct {
		index map[string]map[string]int
		data  []*model.Metric
	}
	type args struct {
		mr *model.MetricRequest
	}
	gauge, err := model.NewMetricRequest(model.TypeGauge, "test", "23")
	require.NoError(t, err)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Metric
		wantErr error
	}{
		{
			name: "create",
			fields: fields{
				index: map[string]map[string]int{},
				data:  []*model.Metric{},
			},
			args: args{
				mr: gauge,
			},
			want:    gauge.Clone(),
			wantErr: nil,
		},
		{
			name: "already exist",
			fields: fields{
				index: map[string]map[string]int{
					model.TypeGauge: {
						"test": 0,
					},
				},
				data: []*model.Metric{gauge.Clone()},
			},
			args: args{
				mr: gauge,
			},
			want:    nil,
			wantErr: model.ErrMetricAlreadyExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := newMemStorageWithDataAndIndex(tt.fields.data, tt.fields.index)
			got, err := ms.Create(context.TODO(), tt.args.mr)
			assert.Equal(t, tt.want, got)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				checkMetricNotSame(t, ms.data[0], got)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_Find(t *testing.T) {
	type fields struct {
		index map[string]map[string]int
		data  []*model.Metric
	}
	type args struct {
		mr *model.MetricRequest
	}
	gauge, err := model.NewMetricRequest(model.TypeGauge, "test", "23")
	require.NoError(t, err)
	counter, err := model.NewMetricRequest(model.TypeCounter, "test", "13")
	require.NoError(t, err)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Metric
		wantErr error
	}{
		{
			name: "found gauge",
			fields: fields{
				index: map[string]map[string]int{
					model.TypeGauge: {
						"test": 0,
					},
				},
				data: []*model.Metric{gauge.Clone()},
			},
			args: args{
				mr: gauge,
			},
			want:    gauge.Clone(),
			wantErr: nil,
		},
		{
			name: "found counter",
			fields: fields{
				index: map[string]map[string]int{
					model.TypeCounter: {
						"test": 0,
					},
				},
				data: []*model.Metric{counter.Clone()},
			},
			args: args{
				mr: counter,
			},
			want:    counter.Clone(),
			wantErr: nil,
		},
		{
			name: "not found",
			fields: fields{
				index: map[string]map[string]int{},
				data:  []*model.Metric{},
			},
			args: args{
				mr: gauge,
			},
			want:    nil,
			wantErr: model.ErrMetricNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := newMemStorageWithDataAndIndex(tt.fields.data, tt.fields.index)
			got, err := ms.Find(context.TODO(), tt.args.mr)
			assert.Equal(t, tt.want, got)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				checkMetricNotSame(t, ms.data[0], got)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_FindAll(t *testing.T) {
	type fields struct {
		index map[string]map[string]int
		data  []*model.Metric
	}
	tests := []struct {
		name   string
		fields fields
		want   []*model.Metric
	}{
		{
			name: "found",
			fields: fields{
				index: map[string]map[string]int{
					model.TypeGauge: {
						"test": 0,
					},
				},
				data: []*model.Metric{
					model.NewMetricGauge("test", 13),
					model.NewMetricCounter("test", 13),
				},
			},
			want: []*model.Metric{
				model.NewMetricGauge("test", 13),
				model.NewMetricCounter("test", 13),
			},
		},
		{
			name: "not found",
			fields: fields{
				index: map[string]map[string]int{},
				data:  []*model.Metric{},
			},
			want: []*model.Metric{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := newMemStorageWithDataAndIndex(tt.fields.data, tt.fields.index)
			got, err := ms.FindAll(context.TODO())
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
			for i := range tt.want {
				checkMetricNotSame(t, ms.data[i], got[i])
			}
		})
	}
}

func TestMemStorage_Update(t *testing.T) {
	type fields struct {
		index map[string]map[string]int
		data  []*model.Metric
	}
	type args struct {
		mr *model.MetricRequest
	}
	gauge, err := model.NewMetricRequest(model.TypeGauge, "test", "23")
	require.NoError(t, err)
	counter, err := model.NewMetricRequest(model.TypeCounter, "test", "13")
	require.NoError(t, err)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Metric
		wantErr error
	}{
		{
			name: "update gauge",
			fields: fields{
				index: map[string]map[string]int{
					model.TypeGauge: {
						"test": 0,
					},
				},
				data: []*model.Metric{
					model.NewMetricGauge("test", 4),
				},
			},
			args: args{
				mr: gauge,
			},
			want:    model.NewMetricGauge("test", 23),
			wantErr: nil,
		},
		{
			name: "update counter",
			fields: fields{
				index: map[string]map[string]int{
					model.TypeCounter: {
						"test": 0,
					},
				},
				data: []*model.Metric{
					model.NewMetricCounter("test", 5),
				},
			},
			args: args{
				mr: counter,
			},
			want:    model.NewMetricCounter("test", 13),
			wantErr: nil,
		},
		{
			name: "not found",
			fields: fields{
				index: map[string]map[string]int{
					model.TypeCounter: {
						"test": 0,
					},
				},
				data: []*model.Metric{
					model.NewMetricCounter("test", 5),
				},
			},
			args: args{
				mr: gauge,
			},
			want:    nil,
			wantErr: model.ErrMetricNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := newMemStorageWithDataAndIndex(tt.fields.data, tt.fields.index)
			got, err := ms.Update(context.TODO(), tt.args.mr)
			assert.Equal(t, tt.want, got)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				checkMetricNotSame(t, ms.data[0], got)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_fill(t *testing.T) {
	type fields struct {
		index map[string]map[string]int
		data  []*model.Metric
	}
	type args struct {
		data []*model.Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "empty",
			fields: fields{
				index: map[string]map[string]int{
					model.TypeGauge: {
						"test": 0,
					},
				},
				data: []*model.Metric{
					model.NewMetricCounter("test", 1),
				},
			},
			args: args{
				data: []*model.Metric{},
			},
		},
		{
			name: "one gauge",
			fields: fields{
				index: map[string]map[string]int{},
				data:  []*model.Metric{},
			},
			args: args{
				data: []*model.Metric{
					model.NewMetricGauge("test1", 1),
				},
			},
		},
		{
			name: "one counter",
			fields: fields{
				index: map[string]map[string]int{},
				data:  []*model.Metric{},
			},
			args: args{
				data: []*model.Metric{
					model.NewMetricCounter("test1", 1),
				},
			},
		},
		{
			name: "replace",
			fields: fields{
				index: map[string]map[string]int{
					model.TypeGauge: {
						"test": 0,
					},
				},
				data: []*model.Metric{
					model.NewMetricCounter("test", 10),
				},
			},
			args: args{
				data: []*model.Metric{
					model.NewMetricGauge("test1", 1),
					model.NewMetricCounter("test2", 2),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := newMemStorageWithDataAndIndex(tt.fields.data, tt.fields.index)
			ms.fill(tt.args.data)
			require.Equal(t, ms.data, tt.args.data)
			for i := range ms.data {
				checkMetricNotSame(t, ms.data[i], tt.args.data[i])
			}
		})
	}
}

func newMemStorageWithDataAndIndex(data []*model.Metric, index map[string]map[string]int) *MemStorage {
	return &MemStorage{
		mux:   new(sync.Mutex),
		index: index,
		data:  data,
	}
}

func checkMetricNotSame(t *testing.T, src *model.Metric, got *model.Metric) {
	t.Helper()
	require.Equal(t, src, got)
	assert.NotSame(t, src, got)
	if src.Value != nil {
		require.NotNil(t, got.Value)
		assert.NotSame(t, src.Value, got.Value)
	} else {
		require.Nil(t, got.Value)
	}
	if src.Delta != nil {
		require.NotNil(t, got.Delta)
		assert.NotSame(t, src.Delta, got.Delta)
	} else {
		require.Nil(t, got.Delta)
	}
}
