package factory

import (
	"context"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository/pgxstorage"
	"github.com/korobkovandrey/runtime-metrics/internal/server/service"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

type Repository interface {
	service.FinderRepository
	service.UpdaterRepository
	service.BatchUpdaterRepository
}

func RepositoryFactory(ctx context.Context, cfg *config.Config, l *logging.ZapLogger) (Repository, error) {
	if cfg.DatabaseDSN != "" {
		ps, err := pgxstorage.NewPGXStorage(ctx, &pgxstorage.Config{
			DSN:         cfg.DatabaseDSN,
			PingTimeout: cfg.DatabasePingTimeout,
			RetryDelays: cfg.RetryDelays,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create pgxstorage: %w", err)
		}
		return ps, nil
	}
	ms := repository.NewMemStorage()
	if cfg.FileStoragePath != "" {
		fs := repository.NewFileStorage(ms, cfg)
		if cfg.Restore {
			if err := fs.Restore(); err != nil {
				return nil, fmt.Errorf("failed to restore: %w", err)
			}
		}
		go fs.Run(ctx, l)
		return fs, nil
	}
	return ms, nil
}
