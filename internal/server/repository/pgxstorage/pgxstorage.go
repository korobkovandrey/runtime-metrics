package pgxstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

type Config struct {
	DSN         string
	PingTimeout time.Duration
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

func (ps *PGXStorage) Find(mr *model.MetricRequest) (*model.Metric, error) {
	row := ps.db.QueryRow(`
		SELECT type, id, value, delta FROM metrics
		WHERE type = $1 AND id = $2 LIMIT 1;`,
		mr.MType, mr.ID,
	)
	m := &model.Metric{}
	err := m.ScanRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrMetricNotFound
		}
		return nil, fmt.Errorf("pxgstorage.Find: %w", err)
	}
	return m, nil
}

func (ps *PGXStorage) FindAll() ([]*model.Metric, error) {
	rows, err := ps.db.Query("SELECT type, id, value, delta FROM metrics ORDER BY type, id;")
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	return scanMetricsFromRows(rows)
}

func (ps *PGXStorage) FindBatch(mrs []*model.MetricRequest) ([]*model.Metric, error) {
	const numColumns = 2
	params := make([]any, 0, len(mrs)*numColumns)
	args := make([]string, 0, len(mrs))
	for i, mr := range mrs {
		k := i*numColumns + 1
		args = append(args, "(type=$"+strconv.Itoa(k)+" AND id=$"+strconv.Itoa(k+1)+")")
		params = append(params, mr.MType, mr.ID)
	}
	rows, err := ps.db.Query(fmt.Sprintf(
		`SELECT type, id, value, delta FROM metrics WHERE %s ORDER BY type, id;`,
		strings.Join(args, " OR ")), params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	return scanMetricsFromRows(rows)
}

func scanMetricsFromRows(rows *sql.Rows) ([]*model.Metric, error) {
	var metrics []*model.Metric
	for rows.Next() {
		m := &model.Metric{}
		if err := m.ScanRow(rows); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		metrics = append(metrics, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}
	return metrics, nil
}

func (ps *PGXStorage) Create(mr *model.MetricRequest) (*model.Metric, error) {
	row := ps.db.QueryRow(`
		INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4)
		RETURNING type, id, value, delta;`,
		mr.MType, mr.ID, mr.Value, mr.Delta,
	)
	m := &model.Metric{}
	err := m.ScanRow(row)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ERROR: duplicate key value violates unique constraint") {
			return nil, model.ErrMetricAlreadyExist
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}
	return m, nil
}

func (ps *PGXStorage) Update(mr *model.MetricRequest) (*model.Metric, error) {
	row := ps.db.QueryRow(`
		INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4)
		ON CONFLICT (type, id) DO UPDATE SET value = EXCLUDED.value, delta = EXCLUDED.delta
		RETURNING type, id, value, delta;`,
		mr.MType, mr.ID, mr.Value, mr.Delta,
	)
	m := &model.Metric{}
	err := m.ScanRow(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}
	return m, nil
}

func (ps *PGXStorage) CreateOrUpdateBatch(mrs []*model.MetricRequest) ([]*model.Metric, error) {
	const numColumns = 4
	params := make([]any, 0, len(mrs)*numColumns)
	args := make([]string, 0, len(mrs))
	for i, mr := range mrs {
		k := i*numColumns + 1
		args = append(args,
			"($"+strconv.Itoa(k)+", $"+strconv.Itoa(k+1)+
				", $"+strconv.Itoa(k+2)+", $"+strconv.Itoa(k+3)+")")
		params = append(params, mr.MType, mr.ID, mr.Value, mr.Delta)
	}
	rows, err := ps.db.Query(fmt.Sprintf(
		`INSERT INTO metrics (type, id, value, delta) VALUES %s 
            ON CONFLICT (type, id) DO UPDATE SET value = EXCLUDED.value, delta = EXCLUDED.delta
            RETURNING type, id, value, delta;`, strings.Join(args, ", ")), params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	return scanMetricsFromRows(rows)
}

//nolint:wrapcheck // ignore
func (ps *PGXStorage) Close() error {
	return ps.db.Close()
}

//nolint:wrapcheck // ignore
func (ps *PGXStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ps.cfg.PingTimeout)
	defer cancel()
	return ps.db.PingContext(ctx)
}
