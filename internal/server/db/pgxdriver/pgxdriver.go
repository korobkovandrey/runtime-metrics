package pgxdriver

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

type DB struct {
	pool *pgxpool.Pool
	cfg  *Config
	l    *logging.ZapLogger
}

type Config struct {
	DSN string
}

func NewDB(ctx context.Context, cfg *Config, l *logging.ZapLogger) (*DB, error) {
	db := &DB{
		cfg: cfg,
		l:   l,
	}
	err := db.init(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewDB: %w", err)
	}
	err = db.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewDB: %w", err)
	}
	return db, nil
}

func (db *DB) init(ctx context.Context) error {
	poolCfg, err := pgxpool.ParseConfig(db.cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse the DSN: %w", err)
	}
	db.pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	return nil
}

//nolint:wrapcheck // ignore
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *DB) Close() {
	db.pool.Close()
}
