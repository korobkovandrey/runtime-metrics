package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	l         *logging.ZapLogger
	isSync    bool
	isChanged bool
}

func NewFileStorage(ms *MemStorage, cfg *config.Config, l *logging.ZapLogger) *FileStorage {
	return &FileStorage{
		MemStorage: ms,
		cfg:        cfg,
		l:          l,
		isSync:     cfg.StoreInterval <= 0,
	}
}

func (fs *FileStorage) Create(mr *model.MetricRequest) (*model.Metric, error) {
	fs.mux.Lock()
	defer fs.mux.Unlock()
	fs.isChanged = true
	m, err := fs.unsafeCreate(mr)
	if err != nil {
		return m, fmt.Errorf("filestorage.Create: %w", err)
	}
	if fs.isSync {
		err = fs.sync(false)
		if err != nil {
			return m, fmt.Errorf("filestorage.Create: %w", err)
		}
	}
	return m, nil
}

func (fs *FileStorage) Update(mr *model.MetricRequest) (*model.Metric, error) {
	fs.mux.Lock()
	defer fs.mux.Unlock()
	fs.isChanged = true
	m, err := fs.unsafeUpdate(mr)
	if err != nil {
		return m, fmt.Errorf("filestorage.Update: %w", err)
	}
	if fs.isSync {
		err = fs.sync(false)
		if err != nil {
			return m, fmt.Errorf("filestorage.Update: %w", err)
		}
	}
	return m, nil
}

func (fs *FileStorage) restore() error {
	if fs.cfg.FileStoragePath == "" {
		return nil
	}
	stat, err := os.Stat(fs.cfg.FileStoragePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("filestorage.restore: %w", err)
		}
		return nil
	}
	if stat.Size() > 0 {
		data, err := os.ReadFile(fs.cfg.FileStoragePath)
		if err != nil {
			return fmt.Errorf("filestorage.restore: %w", err)
		}
		if len(data) > 0 {
			var mrs []*model.Metric
			err = json.Unmarshal(data, &mrs)
			if err != nil {
				return fmt.Errorf("filestorage.restore: %w", err)
			}
			fs.fill(mrs)
		}
	}
	return nil
}

func (fs *FileStorage) sync(safe bool) error {
	if safe {
		fs.mux.Lock()
		defer fs.mux.Unlock()
	}
	if fs.cfg.FileStoragePath == "" {
		return nil
	}
	if !fs.isChanged {
		return nil
	}
	data, err := json.MarshalIndent(fs.unsafeFindAll(), "", "   ")
	if err != nil {
		return fmt.Errorf("filestorage.sync: %w", err)
	}
	const (
		permFlag = 0o600
	)
	err = os.WriteFile(fs.cfg.FileStoragePath, data, permFlag)
	if err != nil {
		return fmt.Errorf("filestorage.sync: %w", err)
	}
	fs.isChanged = false
	return nil
}

func (fs *FileStorage) run(ctx context.Context, l *logging.ZapLogger) {
	if fs.cfg.StoreInterval <= 0 {
		return
	}
	var err error
	t := time.NewTicker(time.Duration(fs.cfg.StoreInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
			if err = fs.sync(true); err != nil {
				l.ErrorCtx(ctx, "filestorage.run", zap.Error(err))
			}
		}
	}
}
