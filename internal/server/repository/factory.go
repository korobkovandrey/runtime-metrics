package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/service"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

func Factory(ctx context.Context, wg *sync.WaitGroup,
	cfg *config.Config, l *logging.ZapLogger) (service.Repository, error) {
	ms := NewMemStorage()
	if cfg.FileStoragePath != "" {
		fs := NewFileStorage(ms, cfg, l)
		if cfg.Restore {
			if err := fs.restore(); err != nil {
				return nil, fmt.Errorf("repository.Factory: %w", err)
			}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()
			l.InfoCtx(ctx, "Syncing filestorage...")
			if err := fs.sync(true); err != nil {
				l.ErrorCtx(ctx, "repository.Factory", zap.Error(err))
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			fs.run(ctx, l)
		}()
		return fs, nil
	}
	return ms, nil
}
