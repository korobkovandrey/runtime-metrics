// Package pgxstorage contains a PostgreSQL storage for metrics.
//
// It uses jackc/pgx library for database operations.
package pgxstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

// Config is the PostgreSQL storage config.
type Config struct {
	DSN         string
	RetryDelays []time.Duration
	PingTimeout time.Duration
}

// PGXStorage is the PostgreSQL storage.
type PGXStorage struct {
	cfg   *Config
	db    *sql.DB
	stmts *statements
}

// NewPGXStorage creates a new PostgreSQL storage.
func NewPGXStorage(ctx context.Context, cfg *Config) (*PGXStorage, error) {
	ps := &PGXStorage{cfg: cfg}
	var err error
	if err = runMigrations(ps.cfg.DSN); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	ps.db, err = sql.Open("pgx", ps.cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	ps.db.SetMaxOpenConns(5) //nolint:mnd // ignore
	ps.db.SetMaxIdleConns(5) //nolint:mnd // ignore
	ps.db.SetConnMaxIdleTime(time.Minute)
	ps.db.SetConnMaxLifetime(time.Minute)
	err = ps.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	ps.stmts, err = ps.prepareStatements(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}
	return ps, nil
}

// Close closes the PostgreSQL storage.
func (ps *PGXStorage) Close() error {
	return ps.db.Close()
}

// Ping pings the PostgreSQL storage.
func (ps *PGXStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ps.cfg.PingTimeout)
	defer cancel()
	return ps.db.PingContext(ctx)
}

// Find returns the metric with the given ID.
func (ps *PGXStorage) Find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	return ps.retryForOne(ctx, mr, ps.find)
}

// FindAll returns all metrics.
func (ps *PGXStorage) FindAll(ctx context.Context) (res []*model.Metric, err error) {
	var e *pgconn.PgError
	for i := 0; ; i++ {
		res, err = ps.findAll(ctx)
		if i == len(ps.cfg.RetryDelays) ||
			err == nil || !errors.As(err, &e) || !pgerrcode.IsConnectionException(e.Code) {
			break
		}
		time.Sleep(ps.cfg.RetryDelays[i])
	}
	return res, err
}

// FindBatch returns the metrics with the given IDs.
func (ps *PGXStorage) FindBatch(ctx context.Context, mrs []*model.MetricRequest) (res []*model.Metric, err error) {
	var e *pgconn.PgError
	for i := 0; ; i++ {
		res, err = ps.findBatch(ctx, mrs)
		if i == len(ps.cfg.RetryDelays) ||
			err == nil || !errors.As(err, &e) || !pgerrcode.IsConnectionException(e.Code) {
			break
		}
		time.Sleep(ps.cfg.RetryDelays[i])
	}
	return res, err
}

// Create creates a new metric.
func (ps *PGXStorage) Create(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	return ps.retryForOne(ctx, mr, ps.create)
}

// Update updates the metric.
func (ps *PGXStorage) Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	return ps.retryForOne(ctx, mr, ps.update)
}

// CreateOrUpdateBatch creates or updates the metrics.
func (ps *PGXStorage) CreateOrUpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
	var e *pgconn.PgError
	var err error
	for i := 0; ; i++ {
		err = ps.createOrUpdateBatch(ctx, mrs)
		if i == len(ps.cfg.RetryDelays) ||
			err == nil || !errors.As(err, &e) || !pgerrcode.IsConnectionException(e.Code) {
			break
		}
		time.Sleep(ps.cfg.RetryDelays[i])
	}
	if err != nil {
		return nil, err
	}
	return ps.FindBatch(ctx, mrs)
}

// retryForOne retries the function for one metric.
func (ps *PGXStorage) retryForOne(ctx context.Context, mr *model.MetricRequest,
	f func(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error)) (m *model.Metric, err error) {
	var e *pgconn.PgError
	for i := 0; ; i++ {
		m, err = f(ctx, mr)
		if i == len(ps.cfg.RetryDelays) ||
			err == nil || !errors.As(err, &e) || !pgerrcode.IsConnectionException(e.Code) {
			break
		}
		time.Sleep(ps.cfg.RetryDelays[i])
	}
	return m, err
}
