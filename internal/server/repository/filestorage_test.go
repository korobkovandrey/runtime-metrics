package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func (f *FileStorage) isChangedF() bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	return f.isChanged
}

func TestFileStorage_Create(t *testing.T) {
	type args struct {
		mr *model.MetricRequest
	}
	gauge, err := model.NewMetricRequest(model.TypeGauge, "test", "23")
	require.NoError(t, err)

	tests := []struct {
		wantErr     error
		cfg         *config.Config
		args        args
		want        *model.Metric
		name        string
		fileContent []*model.Metric
		checkFile   bool
	}{
		{
			name: "create with sync",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				StoreInterval:   0, // Sync mode
				RetryDelays:     []time.Duration{0},
			},
			args: args{
				mr: gauge,
			},
			want:        gauge.Clone(),
			wantErr:     nil,
			checkFile:   true,
			fileContent: []*model.Metric{gauge.Clone()},
		},
		{
			name: "create without sync",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				StoreInterval:   10, // Async mode
				RetryDelays:     []time.Duration{0},
			},
			args: args{
				mr: gauge,
			},
			want:        gauge.Clone(),
			wantErr:     nil,
			checkFile:   false,
			fileContent: nil,
		},
		{
			name: "create with invalid path",
			cfg: &config.Config{
				FileStoragePath: "/invalid/path/metrics.json",
				StoreInterval:   0, // Sync mode
				RetryDelays:     []time.Duration{0},
			},
			args: args{
				mr: gauge,
			},
			want:    gauge.Clone(),
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMemStorage()
			fs := NewFileStorage(ms, tt.cfg)
			got, err := fs.Create(t.Context(), tt.args.mr)
			assert.Equal(t, tt.want, got)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				checkMetricNotSame(t, ms.data[0], got)
			} else {
				require.Error(t, err)
			}

			if tt.checkFile {
				data, err := os.ReadFile(tt.cfg.FileStoragePath)
				require.NoError(t, err)
				var metrics []*model.Metric
				err = json.Unmarshal(data, &metrics)
				require.NoError(t, err)
				assert.Equal(t, tt.fileContent, metrics)
			}
		})
	}
}

func TestFileStorage_Update(t *testing.T) {
	type args struct {
		mr *model.MetricRequest
	}
	gauge, err := model.NewMetricRequest(model.TypeGauge, "test", "23")
	require.NoError(t, err)

	tests := []struct {
		wantErr     error
		cfg         *config.Config
		args        args
		index       map[string]map[string]int
		want        *model.Metric
		name        string
		data        []*model.Metric
		fileContent []*model.Metric
		checkFile   bool
	}{
		{
			name: "update with sync",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				StoreInterval:   0,
				RetryDelays:     []time.Duration{0},
			},
			args: args{
				mr: gauge,
			},
			data: []*model.Metric{model.NewMetricGauge("test", 4)},
			index: map[string]map[string]int{
				model.TypeGauge: {"test": 0},
			},
			want:        model.NewMetricGauge("test", 23),
			wantErr:     nil,
			checkFile:   true,
			fileContent: []*model.Metric{model.NewMetricGauge("test", 23)},
		},
		{
			name: "update without sync",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				StoreInterval:   10,
				RetryDelays:     []time.Duration{0},
			},
			args: args{
				mr: gauge,
			},
			data: []*model.Metric{model.NewMetricGauge("test", 4)},
			index: map[string]map[string]int{
				model.TypeGauge: {"test": 0},
			},
			want:      model.NewMetricGauge("test", 23),
			wantErr:   nil,
			checkFile: false,
		},
		{
			name: "update not found",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				StoreInterval:   0,
				RetryDelays:     []time.Duration{0},
			},
			args: args{
				mr: gauge,
			},
			data:        []*model.Metric{},
			index:       map[string]map[string]int{},
			want:        nil,
			wantErr:     model.ErrMetricNotFound,
			checkFile:   false,
			fileContent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := newMemStorageWithDataAndIndex(tt.data, tt.index)
			fs := NewFileStorage(ms, tt.cfg)
			got, err := fs.Update(t.Context(), tt.args.mr)
			assert.Equal(t, tt.want, got)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				checkMetricNotSame(t, ms.data[0], got)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}

			if tt.checkFile {
				data, err := os.ReadFile(tt.cfg.FileStoragePath)
				require.NoError(t, err)
				var metrics []*model.Metric
				err = json.Unmarshal(data, &metrics)
				require.NoError(t, err)
				assert.Equal(t, tt.fileContent, metrics)
			}
		})
	}
}

func TestFileStorage_CreateOrUpdateBatch(t *testing.T) {
	type args struct {
		mrs []*model.MetricRequest
	}
	gauge1, err := model.NewMetricRequest(model.TypeGauge, "test1", "23")
	require.NoError(t, err)
	gauge2, err := model.NewMetricRequest(model.TypeGauge, "test2", "42")
	require.NoError(t, err)

	tests := []struct {
		wantErr     error
		cfg         *config.Config
		index       map[string]map[string]int
		name        string
		args        args
		data        []*model.Metric
		want        []*model.Metric
		fileContent []*model.Metric
		checkFile   bool
	}{
		{
			name: "create and update with sync",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				StoreInterval:   0,
				RetryDelays:     []time.Duration{0},
			},
			args: args{
				mrs: []*model.MetricRequest{gauge1, gauge2},
			},
			data: []*model.Metric{model.NewMetricGauge("test1", 4)},
			index: map[string]map[string]int{
				model.TypeGauge: {"test1": 0},
			},
			want: []*model.Metric{
				model.NewMetricGauge("test1", 23),
				gauge2.Clone(),
			},
			wantErr:   nil,
			checkFile: true,
			fileContent: []*model.Metric{
				model.NewMetricGauge("test1", 23),
				gauge2.Clone(),
			},
		},
		{
			name: "create and update without sync",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				StoreInterval:   10,
				RetryDelays:     []time.Duration{0},
			},
			args: args{
				mrs: []*model.MetricRequest{gauge1, gauge2},
			},
			data: []*model.Metric{model.NewMetricGauge("test1", 4)},
			index: map[string]map[string]int{
				model.TypeGauge: {"test1": 0},
			},
			want: []*model.Metric{
				model.NewMetricGauge("test1", 23),
				gauge2.Clone(),
			},
			wantErr:   nil,
			checkFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := newMemStorageWithDataAndIndex(tt.data, tt.index)
			fs := NewFileStorage(ms, tt.cfg)
			got, err := fs.CreateOrUpdateBatch(t.Context(), tt.args.mrs)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
			for i := range got {
				checkMetricNotSame(t, ms.data[ms.index[tt.args.mrs[i].MType][tt.args.mrs[i].ID]], got[i])
			}

			if tt.checkFile {
				data, err := os.ReadFile(tt.cfg.FileStoragePath)
				require.NoError(t, err)
				var metrics []*model.Metric
				err = json.Unmarshal(data, &metrics)
				require.NoError(t, err)
				assert.Equal(t, tt.fileContent, metrics)
			}
		})
	}
}

func TestFileStorage_Restore(t *testing.T) {
	gauge, err := model.NewMetricRequest(model.TypeGauge, "test", "23")
	require.NoError(t, err)
	metrics := []*model.Metric{gauge.Clone()}
	data, err := json.MarshalIndent(metrics, "", "   ")
	require.NoError(t, err)

	tests := []struct {
		wantErr   error
		cfg       *config.Config
		wantIndex map[string]map[string]int
		name      string
		fileData  []byte
		wantData  []*model.Metric
	}{
		{
			name: "restore valid data",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				RetryDelays:     []time.Duration{0},
			},
			fileData: data,
			wantData: metrics,
			wantIndex: map[string]map[string]int{
				model.TypeGauge: {"test": 0},
			},
			wantErr: nil,
		},
		{
			name: "restore empty file",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				RetryDelays:     []time.Duration{0},
			},
			fileData:  []byte{},
			wantData:  []*model.Metric{},
			wantIndex: map[string]map[string]int{},
			wantErr:   nil,
		},
		{
			name: "no file",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "nonexistent.json"),
				RetryDelays:     []time.Duration{0},
			},
			fileData:  nil,
			wantData:  []*model.Metric{},
			wantIndex: map[string]map[string]int{},
			wantErr:   nil,
		},
		{
			name: "invalid JSON",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				RetryDelays:     []time.Duration{0},
			},
			fileData:  []byte("invalid json"),
			wantData:  nil,
			wantIndex: map[string]map[string]int{},
			wantErr:   assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fileData != nil {
				err := os.WriteFile(tt.cfg.FileStoragePath, tt.fileData, 0o600)
				require.NoError(t, err)
			}
			ms := NewMemStorage()
			fs := NewFileStorage(ms, tt.cfg)
			err := fs.Restore()
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantData, ms.data)
				assert.Equal(t, tt.wantIndex, ms.index)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFileStorage_sync(t *testing.T) {
	gauge, err := model.NewMetricRequest(model.TypeGauge, "test", "23")
	require.NoError(t, err)
	data := []*model.Metric{gauge.Clone()}

	tests := []struct {
		wantErr   error
		cfg       *config.Config
		index     map[string]map[string]int
		name      string
		data      []*model.Metric
		isChanged bool
		safe      bool
		tryRetry  bool
	}{
		{
			name: "sync with data",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				RetryDelays:     []time.Duration{0},
			},
			data:      data,
			index:     map[string]map[string]int{model.TypeGauge: {"test": 0}},
			isChanged: true,
			safe:      true,
			tryRetry:  true,
			wantErr:   nil,
		},
		{
			name: "sync unchanged",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
				RetryDelays:     []time.Duration{0},
			},
			data:      data,
			index:     map[string]map[string]int{model.TypeGauge: {"test": 0}},
			isChanged: false,
			safe:      true,
			tryRetry:  true,
			wantErr:   nil,
		},
		{
			name: "sync empty path",
			cfg: &config.Config{
				FileStoragePath: "",
				RetryDelays:     []time.Duration{0},
			},
			data:      data,
			index:     map[string]map[string]int{model.TypeGauge: {"test": 0}},
			isChanged: true,
			safe:      true,
			tryRetry:  true,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := newMemStorageWithDataAndIndex(tt.data, tt.index)
			fs := NewFileStorage(ms, tt.cfg)
			fs.isChanged = tt.isChanged
			err := fs.sync(tt.safe, tt.tryRetry)
			assert.NoError(t, err)

			if tt.isChanged && tt.cfg.FileStoragePath != "" {
				fileData, err := os.ReadFile(tt.cfg.FileStoragePath)
				require.NoError(t, err)
				var metrics []*model.Metric
				err = json.Unmarshal(fileData, &metrics)
				require.NoError(t, err)
				assert.Equal(t, tt.data, metrics)
			}
		})
	}
}

func TestFileStorage_Run(t *testing.T) {
	gauge, err := model.NewMetricRequest(model.TypeGauge, "test", "23")
	require.NoError(t, err)
	data := []*model.Metric{gauge.Clone()}
	tim := time.Now().String()

	tests := []struct {
		cfg       *config.Config
		index     map[string]map[string]int
		name      string
		data      []*model.Metric
		isChanged bool
		wantFile  bool
	}{
		{
			name: "run with sync",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics"+tim+".json"),
				StoreInterval:   1,
				RetryDelays:     []time.Duration{0},
			},
			data:      data,
			index:     map[string]map[string]int{model.TypeGauge: {"test": 23}},
			isChanged: true,
			wantFile:  true,
		},
		{
			name: "run with no sync",
			cfg: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "metrics"+tim+".json"),
				StoreInterval:   0,
				RetryDelays:     []time.Duration{0},
			},
			data:      data,
			index:     map[string]map[string]int{model.TypeGauge: {"test": 0}},
			isChanged: true,
			wantFile:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
			defer cancel()

			ms := newMemStorageWithDataAndIndex(tt.data, tt.index)
			fs := NewFileStorage(ms, tt.cfg)
			fs.isChanged = tt.isChanged
			logger, err := logging.NewZapLogger(zapcore.DebugLevel)
			require.NoError(t, err)

			go fs.Run(ctx, logger)

			// Wait for potential sync
			time.Sleep(2*time.Duration(tt.cfg.StoreInterval)*time.Second + 10*time.Millisecond)

			if tt.wantFile {
				fileData, err := os.ReadFile(tt.cfg.FileStoragePath)
				require.NoError(t, err)
				var metrics []*model.Metric
				err = json.Unmarshal(fileData, &metrics)
				require.NoError(t, err)
				assert.Equal(t, tt.data, metrics)
				assert.False(t, fs.isChangedF())
			} else {
				_, err := os.Stat(tt.cfg.FileStoragePath)
				assert.Error(t, err)
			}
		})
	}
}
