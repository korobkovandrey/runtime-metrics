package pgxstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

func (ps *PGXStorage) find(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	row := ps.stmts.findOneStmt.QueryRowContext(ctx, mr.MType, mr.ID)
	m := &model.Metric{}
	err := m.ScanRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("%w: %w", model.ErrMetricNotFound, err)
		}
		return nil, fmt.Errorf("failed to find metric with type=%s and id=%s: %w", mr.MType, mr.ID, err)
	}
	return m, nil
}

func (ps *PGXStorage) findAll(ctx context.Context) ([]*model.Metric, error) {
	rows, err := ps.stmts.findAllStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	return scanMetricsFromRows(rows)
}

func (ps *PGXStorage) findBatch(ctx context.Context, mrs []*model.MetricRequest) ([]*model.Metric, error) {
	q, params := makeFindBatchQuery(mrs)
	rows, err := ps.db.QueryContext(ctx, q, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	return scanMetricsFromRows(rows)
}

func (ps *PGXStorage) create(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	row := ps.stmts.createReturningStmt.QueryRowContext(ctx, mr.MType, mr.ID, mr.Value, mr.Delta)
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
	row := ps.stmts.updateReturningStmt.QueryRowContext(ctx, mr.Value, mr.Delta, mr.MType, mr.ID)
	m := &model.Metric{}
	err := m.ScanRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("%w: %w", model.ErrMetricNotFound, err)
		}
		return nil, fmt.Errorf("failed to update metric with type=%s and id=%s: %w", mr.MType, mr.ID, err)
	}
	return m, nil
}

func (ps *PGXStorage) createOrUpdateBatch(ctx context.Context, mrs []*model.MetricRequest) error {
	tx, err := ps.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, mr := range mrs {
		//nolint:sqlclosecheck // ignore
		_, err = tx.StmtContext(ctx, ps.stmts.upsertStmt).ExecContext(ctx, mr.MType, mr.ID, mr.Value, mr.Delta)
		if err != nil {
			return fmt.Errorf("failed to query: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
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
