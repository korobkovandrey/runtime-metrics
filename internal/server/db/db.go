package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/db/pgxdriver"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

//go:generate mockgen -source=db.go -destination=../mocks/db.go -package=mocks
type DB interface {
	Ping(ctx context.Context) error
	Close()
}

type Config struct {
	PGXDriver *pgxdriver.Config
}

var (
	ErrNoDBConfig = errors.New("no db config")
)

func Factory(ctx context.Context, cfg *Config, l *logging.ZapLogger) (DB, error) {
	if cfg.PGXDriver != nil {
		db, err := pgxdriver.NewDB(ctx, cfg.PGXDriver, l)
		if err != nil {
			return nil, fmt.Errorf("db.Factory: %w", err)
		}
		return db, nil
	}
	return nil, fmt.Errorf("db.Factory: %w", ErrNoDBConfig)
}
