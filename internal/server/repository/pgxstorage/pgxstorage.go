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

type Config struct {
	DSN         string
	PingTimeout time.Duration
	RetryDelays []time.Duration
}

type PGXStorage struct {
	cfg *Config
	db  *sql.DB
}

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
	err = ps.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return ps, nil
}

func (ps *PGXStorage) Close() error {
	return ps.db.Close()
}

func (ps *PGXStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ps.cfg.PingTimeout)
	defer cancel()
	return ps.db.PingContext(ctx)
}

func (ps *PGXStorage) Find(mr *model.MetricRequest) (*model.Metric, error) {
	return ps.retryForOne(mr, ps.find)
}

func (ps *PGXStorage) FindAll() (res []*model.Metric, err error) {
	var e *pgconn.PgError
	for i := 0; ; i++ {
		res, err = ps.findAll()
		if i == len(ps.cfg.RetryDelays) ||
			err == nil || !errors.As(err, &e) || !pgerrcode.IsConnectionException(e.Code) {
			break
		}
		time.Sleep(ps.cfg.RetryDelays[i])
	}
	return res, err
}

func (ps *PGXStorage) FindBatch(mrs []*model.MetricRequest) ([]*model.Metric, error) {
	return ps.retryForOneMany(mrs, ps.findBatch)
}

func (ps *PGXStorage) Create(mr *model.MetricRequest) (*model.Metric, error) {
	return ps.retryForOne(mr, ps.create)
}

func (ps *PGXStorage) Update(mr *model.MetricRequest) (*model.Metric, error) {
	return ps.retryForOne(mr, ps.update)
}

func (ps *PGXStorage) CreateOrUpdateBatch(mrs []*model.MetricRequest) ([]*model.Metric, error) {
	return ps.retryForOneMany(mrs, ps.createOrUpdateBatch)
}

func (ps *PGXStorage) retryForOne(mr *model.MetricRequest,
	f func(mr *model.MetricRequest) (*model.Metric, error)) (m *model.Metric, err error) {
	var e *pgconn.PgError
	for i := 0; ; i++ {
		m, err = f(mr)
		if i == len(ps.cfg.RetryDelays) ||
			err == nil || !errors.As(err, &e) || !pgerrcode.IsConnectionException(e.Code) {
			break
		}
		time.Sleep(ps.cfg.RetryDelays[i])
	}
	return m, err
}

func (ps *PGXStorage) retryForOneMany(mrs []*model.MetricRequest,
	f func(mrs []*model.MetricRequest) ([]*model.Metric, error)) (res []*model.Metric, err error) {
	var e *pgconn.PgError
	for i := 0; ; i++ {
		res, err = f(mrs)
		if i == len(ps.cfg.RetryDelays) ||
			err == nil || !errors.As(err, &e) || !pgerrcode.IsConnectionException(e.Code) {
			break
		}
		time.Sleep(ps.cfg.RetryDelays[i])
	}
	return res, err
}
