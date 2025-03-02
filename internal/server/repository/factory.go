package repository

import (
	"context"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/service"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

func Factory(ctx context.Context, cfg *config.Config, l *logging.ZapLogger) (service.Repository, error) {
	ms := NewMemStorage()
	if cfg.FileStoragePath != "" {
		fs := NewFileStorage(ms, cfg, l)
		if cfg.Restore {
			if err := fs.restore(); err != nil {
				return nil, fmt.Errorf("repository.Factory: %w", err)
			}
		}
		go fs.run(ctx, l)
		return fs, nil
	}
	return ms, nil
}
