package repository

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository/pgxstorage"
	"github.com/korobkovandrey/runtime-metrics/internal/server/service"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

//go:generate mockgen -source=factory.go -destination=../mocks/repository.go -package=mocks

type Closer interface {
	Close() error
}

type Pinger interface {
	driver.Pinger
}

type Repository interface {
	service.Repository
}

func Factory(ctx context.Context, cfg *config.Config, l *logging.ZapLogger) (Repository, Closer, Pinger, error) {
	if cfg.DatabaseDSN != "" {
		ps, err := pgxstorage.NewPGXStorage(ctx, &pgxstorage.Config{
			DSN:         cfg.DatabaseDSN,
			PingTimeout: cfg.DatabasePingTimeout,
			RetryDelays: cfg.RetryDelays,
		})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("repository.Factory: %w", err)
		}
		return ps, ps, ps, nil
	}
	ms := NewMemStorage()
	if cfg.FileStoragePath != "" {
		fs := NewFileStorage(ms, cfg)
		if cfg.Restore {
			if err := fs.restore(); err != nil {
				return nil, nil, nil, fmt.Errorf("repository.Factory: %w", err)
			}
		}
		go fs.run(ctx, l)
		return fs, fs, nil, nil
	}
	return ms, nil, nil, nil
}
