package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

type FileStorage struct {
	*MemStorage
	cfg       *config.Config
	isSync    bool
	isChanged bool
}

func NewFileStorage(ms *MemStorage, cfg *config.Config) *FileStorage {
	return &FileStorage{
		MemStorage: ms,
		cfg:        cfg,
		isSync:     cfg.StoreInterval <= 0,
	}
}

func (f *FileStorage) Create(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.isChanged = true
	m, err := f.unsafeCreate(mr)
	if err != nil {
		return m, fmt.Errorf("filestorage.Create: %w", err)
	}
	if f.isSync {
		err = f.sync(false, true)
		if err != nil {
			return m, fmt.Errorf("filestorage.Create: %w", err)
		}
	}
	return m, nil
}

func (f *FileStorage) Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.isChanged = true
	m, err := f.unsafeUpdate(mr)
	if err != nil {
		return m, fmt.Errorf("filestorage.Update: %w", err)
	}
	if f.isSync {
		err = f.sync(false, true)
		if err != nil {
			return m, fmt.Errorf("filestorage.Update: %w", err)
		}
	}
	return m, nil
}

func (f *FileStorage) CreateOrUpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.isChanged = true
	res, err := f.unsafeCreateOrUpdateBatch(mrs)
	if err != nil {
		return res, fmt.Errorf("filestorage.UpdateBatch: %w", err)
	}
	if f.isSync {
		err = f.sync(false, true)
		if err != nil {
			return res, fmt.Errorf("filestorage.UpdateBatch: %w", err)
		}
	}
	return res, nil
}

func (f *FileStorage) Close() error {
	return f.sync(true, false)
}

func (f *FileStorage) restore() error {
	if f.cfg.FileStoragePath == "" {
		return nil
	}
	stat, err := os.Stat(f.cfg.FileStoragePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("filestorage.restore: %w", err)
		}
		return nil
	}
	if stat.Size() > 0 {
		var data []byte
		for i := 0; ; i++ {
			data, err = os.ReadFile(f.cfg.FileStoragePath)
			if i == len(f.cfg.RetryDelays) || err == nil || !errors.Is(err, fs.ErrPermission) {
				break
			}
			time.Sleep(f.cfg.RetryDelays[i])
		}
		if err != nil {
			return fmt.Errorf("filestorage.restore: %w", err)
		}
		if len(data) > 0 {
			var mrs []*model.Metric
			err = json.Unmarshal(data, &mrs)
			if err != nil {
				return fmt.Errorf("filestorage.restore: %w", err)
			}
			f.fill(mrs)
		}
	}
	return nil
}

func (f *FileStorage) sync(safe, tryRetry bool) error {
	if safe {
		f.mux.Lock()
		defer f.mux.Unlock()
	}
	if f.cfg.FileStoragePath == "" {
		return nil
	}
	if !f.isChanged {
		return nil
	}
	data, err := json.MarshalIndent(f.unsafeFindAll(), "", "   ")
	if err != nil {
		return fmt.Errorf("filestorage.sync: %w", err)
	}
	const (
		permFlag = 0o600
	)

	if tryRetry {
		for i := 0; ; i++ {
			err = os.WriteFile(f.cfg.FileStoragePath, data, permFlag)
			if i == len(f.cfg.RetryDelays) || err == nil || !errors.Is(err, fs.ErrPermission) {
				break
			}
			time.Sleep(f.cfg.RetryDelays[i])
		}
	} else {
		err = os.WriteFile(f.cfg.FileStoragePath, data, permFlag)
	}

	if err != nil {
		return fmt.Errorf("filestorage.sync: %w", err)
	}
	f.isChanged = false
	return nil
}

func (f *FileStorage) run(ctx context.Context, l *logging.ZapLogger) {
	if f.cfg.StoreInterval <= 0 {
		return
	}
	var err error
	t := time.NewTicker(time.Duration(f.cfg.StoreInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
			if err = f.sync(true, false); err != nil {
				l.ErrorCtx(ctx, "filestorage.run", zap.Error(err))
			}
		}
	}
}
