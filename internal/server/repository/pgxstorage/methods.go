package pgxstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

func (ps *PGXStorage) find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	row := ps.db.QueryRowContext(ctx, `
		SELECT type, id, value, delta FROM metrics
		WHERE type = $1 AND id = $2 LIMIT 1;`,
		mr.MType, mr.ID,
	)
	m := &model.Metric{}
	err := m.ScanRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = model.ErrMetricNotFound
		}
		return nil, fmt.Errorf("failed to find metric with type=%s and id=%s: %w", mr.MType, mr.ID, err)
	}
	return m, nil
}

func (ps *PGXStorage) findAll(ctx context.Context) ([]*model.Metric, error) {
	rows, err := ps.db.QueryContext(ctx, "SELECT type, id, value, delta FROM metrics ORDER BY type, id;")
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	return scanMetricsFromRows(rows)
}

func (ps *PGXStorage) findBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
	const numColumns = 2
	params := make([]any, 0, len(mrs)*numColumns)
	args := make([]string, 0, len(mrs))
	for i, mr := range mrs {
		k := i*numColumns + 1
		args = append(args, "(type=$"+strconv.Itoa(k)+" AND id=$"+strconv.Itoa(k+1)+")")
		params = append(params, mr.MType, mr.ID)
	}
	rows, err := ps.db.QueryContext(ctx, fmt.Sprintf(
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

func (ps *PGXStorage) create(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	row := ps.db.QueryRowContext(ctx, `
		INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4)
		RETURNING type, id, value, delta;`,
		mr.MType, mr.ID, mr.Value, mr.Delta,
	)
	m := &model.Metric{}
	err := m.ScanRow(row)
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && pgerrcode.IsIntegrityConstraintViolation(e.Code) {
			return nil, model.ErrMetricAlreadyExist
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}
	return m, nil
}

func (ps *PGXStorage) update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	row := ps.db.QueryRowContext(ctx, `
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

func (ps *PGXStorage) createOrUpdateBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
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
	rows, err := ps.db.QueryContext(ctx, fmt.Sprintf(
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
