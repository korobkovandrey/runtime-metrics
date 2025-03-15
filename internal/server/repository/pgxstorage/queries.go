package pgxstorage

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

const (
	findOneQuery         = "SELECT type, id, value, delta FROM metrics WHERE type = $1 AND id = $2 LIMIT 1;"
	findAllQuery         = "SELECT type, id, value, delta FROM metrics ORDER BY type, id;"
	findBatchQueryTpl    = "SELECT type, id, value, delta FROM metrics WHERE %s ORDER BY type, id;"
	createReturningQuery = "INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4) RETURNING type, id, value, delta;"
	updateReturningQuery = "UPDATE metrics SET value = $1, delta = $2 WHERE type = $3 AND id = $4 RETURNING type, id, value, delta;"
	upsertQuery          = `INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4) ON CONFLICT (type, id) 
    DO UPDATE SET value = EXCLUDED.value, delta = EXCLUDED.delta;`
)

func makeFindBatchQuery(mrs []*model.MetricRequest) (q string, params []any) {
	const numColumns = 2
	params = make([]any, 0, len(mrs)*numColumns)
	args := make([]string, 0, len(mrs))
	for i, mr := range mrs {
		k := i*numColumns + 1
		args = append(args, "(type=$"+strconv.Itoa(k)+" AND id=$"+strconv.Itoa(k+1)+")")
		params = append(params, mr.MType, mr.ID)
	}
	return fmt.Sprintf(findBatchQueryTpl, strings.Join(args, " OR ")), params
}

type statements struct {
	findOneStmt         *sql.Stmt
	findAllStmt         *sql.Stmt
	createReturningStmt *sql.Stmt
	updateReturningStmt *sql.Stmt
	upsertStmt          *sql.Stmt
}

func (ps *PGXStorage) prepareStatements(ctx context.Context) (st *statements, err error) {
	st = &statements{}
	st.findOneStmt, err = ps.db.PrepareContext(ctx, findOneQuery)
	if err != nil {
		return st, fmt.Errorf("failed to prepare findOneQuery: %w", err)
	}
	st.findAllStmt, err = ps.db.PrepareContext(ctx, findAllQuery)
	if err != nil {
		return st, fmt.Errorf("failed to prepare findAllQuery: %w", err)
	}
	st.createReturningStmt, err = ps.db.PrepareContext(ctx, createReturningQuery)
	if err != nil {
		return st, fmt.Errorf("failed to prepare createReturningQuery: %w", err)
	}
	st.updateReturningStmt, err = ps.db.PrepareContext(ctx, updateReturningQuery)
	if err != nil {
		return st, fmt.Errorf("failed to prepare updateReturningQuery: %w", err)
	}
	st.upsertStmt, err = ps.db.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return st, fmt.Errorf("failed to prepare upsertQuery: %w", err)
	}
	return
}
