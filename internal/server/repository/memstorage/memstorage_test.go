package memstorage

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server/adapter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestNewMemStorage(t *testing.T) {
	s := NewMemStorage()
	assert.IsType(t, s, &MemStorage{})
	assert.Implements(t, (*adapter.Repository)(nil), s)
}

func TestMemStorage_AddType(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "add type",
			args: args{
				t: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage()
			s.AddType(tt.args.t)
			require.Contains(t, s.data, tt.args.t)
			assert.Equal(t, s.data[tt.args.t], map[string]any{})
		})
	}
}

func TestMemStorage_Set(t *testing.T) {
	type args struct {
		t      string
		name   string
		values []any
	}
	checkIntValues := []any{1, 2, 3, 4, 5}
	checkFloatValues := []any{0.1, 0.2, 0.3, 0.4, 0.5}
	tests := []struct {
		name       string
		args       args
		wantValues []any
	}{
		{
			name: "test1 int",
			args: args{
				t:      "int",
				name:   "test1",
				values: checkIntValues,
			},
			wantValues: checkIntValues,
		},
		{
			name: "test2 int",
			args: args{
				t:      "int",
				name:   "test2",
				values: checkIntValues,
			},
			wantValues: checkIntValues,
		},
		{
			name: "test1 float",
			args: args{
				t:      "float",
				name:   "test1",
				values: checkFloatValues,
			},
			wantValues: checkFloatValues,
		},
		{
			name: "test2 float",
			args: args{
				t:      "float",
				name:   "test2",
				values: checkFloatValues,
			},
			wantValues: checkFloatValues,
		},
	}
	s := NewMemStorage()
	//nolint:dupl // it`s test!
	for _, tt := range tests {
		s.AddType(tt.args.t)
		require.Contains(t, s.data, tt.args.t)
		for i, v := range tt.args.values {
			s.Set(tt.args.t, tt.args.name, v)
			require.Contains(t, s.data[tt.args.t], tt.args.name)
			assert.Equal(t, s.data[tt.args.t][tt.args.name], tt.wantValues[i])
		}
	}
	assert.Equal(t, s.data, map[string]map[string]any{
		"int": {
			"test1": checkIntValues[len(checkIntValues)-1],
			"test2": checkIntValues[len(checkIntValues)-1],
		},
		"float": {
			"test1": checkFloatValues[len(checkFloatValues)-1],
			"test2": checkFloatValues[len(checkFloatValues)-1],
		},
	})
}

func TestMemStorage_IncrInt64(t *testing.T) {
	type args struct {
		t      string
		name   string
		values []int64
	}
	checkIntValues := []int64{1, 2, 3, 4, 5}
	wantIntValues := []int64{1, 3, 6, 10, 15}
	tests := []struct {
		name       string
		args       args
		wantValues []int64
	}{
		{
			name: "test1 int",
			args: args{
				t:      "int",
				name:   "test1",
				values: checkIntValues,
			},
			wantValues: wantIntValues,
		},
		{
			name: "test2 int",
			args: args{
				t:      "int",
				name:   "test2",
				values: checkIntValues,
			},
			wantValues: wantIntValues,
		},
	}
	s := NewMemStorage()
	//nolint:dupl // it`s test!
	for _, tt := range tests {
		s.AddType(tt.args.t)
		require.Contains(t, s.data, tt.args.t)
		for i, v := range tt.args.values {
			s.IncrInt64(tt.args.t, tt.args.name, v)
			require.Contains(t, s.data[tt.args.t], tt.args.name)
			assert.Equal(t, s.data[tt.args.t][tt.args.name], tt.wantValues[i])
		}
	}
	assert.Equal(t, s.data, map[string]map[string]any{
		"int": {
			"test1": wantIntValues[len(wantIntValues)-1],
			"test2": wantIntValues[len(wantIntValues)-1],
		},
	})
}

func TestMemStorage_Get(t *testing.T) {
	type args struct {
		t    string
		name string
	}
	data := map[string]map[string]any{
		"int": {
			"test1": 1,
		},
		"float": {
			"test1": 0.1,
		},
	}

	tests := []struct {
		name      string
		args      args
		wantValue any
		wantOk    bool
	}{

		{
			name: "not exists type",
			args: args{
				t:    "not_exists",
				name: "not_exists",
			},
			wantValue: nil,
			wantOk:    false,
		},
		{
			name: "not exists value",
			args: args{
				t:    "int",
				name: "not_exists",
			},
			wantValue: nil,
			wantOk:    false,
		},
		{
			name: "int value",
			args: args{
				t:    "int",
				name: "test1",
			},
			wantValue: 1,
			wantOk:    true,
		},
		{
			name: "float value",
			args: args{
				t:    "float",
				name: "test1",
			},
			wantValue: 0.1,
			wantOk:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage()
			s.data = data
			gotValue, gotOk := s.Get(tt.args.t, tt.args.name)
			assert.Equal(t, gotValue, tt.wantValue, "MemStorage.Get() gotValue = %v, want %v", gotValue, tt.wantValue)
			assert.Equal(t, gotOk, tt.wantOk, "MemStorage.Get() gotOk = %v, want %v", gotOk, tt.wantOk)
		})
	}
}
