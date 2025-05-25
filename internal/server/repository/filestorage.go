// Package repository provides a file storage implementation for metrics.
//
// The FileStorage struct embeds MemStorage and adds functionality for
// persisting metrics to a file. It supports creating, updating, and
// restoring metrics from a file. The storage can be synchronized
// periodically or on-demand.
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
)

// FileStorage is a file storage for metrics.
type FileStorage struct {
	*MemStorage
	cfg       *config.Config
	isSync    bool
	isChanged bool
}

// NewFileStorage creates a new file storage.
func NewFileStorage(ms *MemStorage, cfg *config.Config) *FileStorage {
	return &FileStorage{
		MemStorage: ms,
		cfg:        cfg,
		isSync:     cfg.StoreInterval <= 0,
	}
}

// Create creates a new metric.
func (f *FileStorage) Create(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.isChanged = true
	m, err := f.unsafeCreate(mr)
	if err != nil {
		return m, fmt.Errorf("failed to create metric: %w", err)
	}
	if f.isSync {
		err = f.sync(false, true)
		if err != nil {
			return m, fmt.Errorf("failed to sync: %w", err)
		}
	}
	return m, nil
}

// Update updates an existing metric.
func (f *FileStorage) Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.isChanged = true
	m, err := f.unsafeUpdate(mr)
	if err != nil {
		return m, fmt.Errorf("failed to update metric: %w", err)
	}
	if f.isSync {
		err = f.sync(false, true)
		if err != nil {
			return m, fmt.Errorf("failed to sync: %w", err)
		}
	}
	return m, nil
}

// CreateOrUpdateBatch creates or updates a batch of metrics.
func (f *FileStorage) CreateOrUpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.isChanged = true
	res, err := f.unsafeCreateOrUpdateBatch(mrs)
	if err != nil {
		return res, fmt.Errorf("failed to create or update metrics: %w", err)
	}
	if f.isSync {
		err = f.sync(false, true)
		if err != nil {
			return res, fmt.Errorf("failed to sync: %w", err)
		}
	}
	return res, nil
}

// Close closes the file storage.
func (f *FileStorage) Close() error {
	return f.sync(true, false)
}

// Restore restores the file storage.
func (f *FileStorage) Restore() error {
	if f.cfg.FileStoragePath == "" {
		return nil
	}
	stat, err := os.Stat(f.cfg.FileStoragePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to stat file: %w", err)
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
			return fmt.Errorf("failed to read file: %w", err)
		}
		if len(data) > 0 {
			var mrs []*model.Metric
			err = json.Unmarshal(data, &mrs)
			if err != nil {
				return fmt.Errorf("failed to unmarshal file: %w", err)
			}
			f.fill(mrs)
		}
	}
	return nil
}

// sync syncs the file storage.
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
		return fmt.Errorf("failed to marshal data: %w", err)
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
		return fmt.Errorf("failed to write file: %w", err)
	}
	f.isChanged = false
	return nil
}

// Run runs the file storage.
func (f *FileStorage) Run(ctx context.Context, l *logging.ZapLogger) {
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
				l.ErrorCtx(ctx, fmt.Errorf("failed to sync: %w", err).Error())
			}
		}
	}
}
