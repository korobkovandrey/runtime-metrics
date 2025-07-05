package pgxstorage

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dbMockPGXStorage(t *testing.T) (*PGXStorage, sqlmock.Sqlmock) {
	t.Helper()
	opt := sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual)
	db, mock, err := sqlmock.New(opt, sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	s := &PGXStorage{
		cfg: &Config{
			RetryDelays: []time.Duration{time.Second},
			PingTimeout: time.Second,
		},
		db: db,
	}
	mock.ExpectPrepare("SELECT type, id, value, delta FROM metrics WHERE type = $1 AND id = $2 LIMIT 1;")
	mock.ExpectPrepare("SELECT type, id, value, delta FROM metrics ORDER BY type, id;")
	mock.ExpectPrepare("INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4) RETURNING type, id, value, delta;")
	mock.ExpectPrepare("UPDATE metrics SET value = $1, delta = $2 WHERE type = $3 AND id = $4 RETURNING type, id, value, delta;")
	mock.ExpectPrepare("INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4) ON CONFLICT (type, id) DO UPDATE SET value = EXCLUDED.value, delta = EXCLUDED.delta;")
	s.stmts, err = s.prepareStatements(t.Context())
	require.NoError(t, err)
	return s, mock
}

func TestNewPGXStorage(t *testing.T) {
	_, err := NewPGXStorage(t.Context(), &Config{})
	require.Error(t, err)
}

func TestPGXStorage_Ping(t *testing.T) {
	s, mock := dbMockPGXStorage(t)
	defer func() {
		assert.NoError(t, s.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	}()
	mock.ExpectPing()
	assert.NoError(t, s.Ping(t.Context()))
	mock.ExpectClose()
}

func TestPGXStorage_Find(t *testing.T) {
	s, mock := dbMockPGXStorage(t)
	defer func() {
		assert.NoError(t, s.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	}()
	t.Run("ok", func(t *testing.T) {
		mock.ExpectQuery("SELECT type, id, value, delta FROM metrics WHERE type = $1 AND id = $2 LIMIT 1;").
			WithArgs("counter", "PollCount").
			WillReturnRows(sqlmock.NewRows([]string{"type", "id", "value", "delta"}).AddRow("counter", "PollCount", nil, 10))
		m, err := s.Find(t.Context(), model.NewMetricCounter("PollCount", 1).ToRequest())
		require.NoError(t, err)
		assert.Equal(t, m, model.NewMetricCounter("PollCount", 10))
	})
	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT type, id, value, delta FROM metrics WHERE type = $1 AND id = $2 LIMIT 1;").
			WithArgs("counter", "PollCount").
			WillReturnRows(sqlmock.NewRows([]string{"type", "id", "value", "delta"}))
		m, err := s.Find(t.Context(), model.NewMetricCounter("PollCount", 1).ToRequest())
		require.Error(t, err)
		assert.ErrorIs(t, err, model.ErrMetricNotFound)
		assert.Nil(t, m)
	})
	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery("SELECT type, id, value, delta FROM metrics WHERE type = $1 AND id = $2 LIMIT 1;").
			WithArgs("counter", "PollCount").
			WillReturnError(errors.New("error"))
		m, err := s.Find(t.Context(), model.NewMetricCounter("PollCount", 1).ToRequest())
		require.Error(t, err)
		assert.NotErrorIs(t, err, model.ErrMetricNotFound)
		assert.Nil(t, m)
	})
	mock.ExpectClose()
}

func TestPGXStorage_FindAll(t *testing.T) {
	s, mock := dbMockPGXStorage(t)
	defer func() {
		assert.NoError(t, s.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	}()
	t.Run("ok", func(t *testing.T) {
		mock.ExpectQuery("SELECT type, id, value, delta FROM metrics ORDER BY type, id;").
			WillReturnRows(sqlmock.NewRows([]string{"type", "id", "value", "delta"}).AddRow("counter", "PollCount", nil, 10))
		m, err := s.FindAll(t.Context())
		require.NoError(t, err)
		assert.Equal(t, m, []*model.Metric{
			model.NewMetricCounter("PollCount", 10),
		})
	})
	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery("SELECT type, id, value, delta FROM metrics ORDER BY type, id;").
			WillReturnError(errors.New("error"))
		m, err := s.FindAll(t.Context())
		require.Error(t, err)
		assert.Nil(t, m)
	})
	mock.ExpectClose()
}

func TestPGXStorage_FindBatch(t *testing.T) {
	s, mock := dbMockPGXStorage(t)
	defer func() {
		assert.NoError(t, s.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	}()
	t.Run("ok", func(t *testing.T) {
		mock.ExpectQuery("SELECT type, id, value, delta FROM metrics WHERE (type=$1 AND id=$2) ORDER BY type, id;").
			WithArgs("counter", "PollCount").
			WillReturnRows(sqlmock.NewRows([]string{"type", "id", "value", "delta"}).AddRow("counter", "PollCount", nil, 10))
		m, err := s.FindBatch(t.Context(), []*model.MetricRequest{model.NewMetricCounter("PollCount", 1).ToRequest()})
		require.NoError(t, err)
		assert.Equal(t, m, []*model.Metric{
			model.NewMetricCounter("PollCount", 10),
		})
	})
	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery("SELECT type, id, value, delta FROM metrics WHERE (type=$1 AND id=$2) ORDER BY type, id;").
			WithArgs("counter", "PollCount").
			WillReturnError(errors.New("error"))
		m, err := s.FindBatch(t.Context(), []*model.MetricRequest{model.NewMetricCounter("PollCount", 1).ToRequest()})
		require.Error(t, err)
		assert.Nil(t, m)
	})
	mock.ExpectClose()
}

func TestPGXStorage_Create(t *testing.T) {
	s, mock := dbMockPGXStorage(t)
	defer func() {
		assert.NoError(t, s.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	}()
	t.Run("ok", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4) RETURNING type, id, value, delta;").
			WithArgs("counter", "PollCount", nil, 10).
			WillReturnRows(sqlmock.NewRows([]string{"type", "id", "value", "delta"}).AddRow("counter", "PollCount", nil, 10))
		m, err := s.Create(t.Context(), model.NewMetricCounter("PollCount", 10).ToRequest())
		require.NoError(t, err)
		assert.Equal(t, m, model.NewMetricCounter("PollCount", 10))
	})
	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4) RETURNING type, id, value, delta;").
			WithArgs("counter", "PollCount", nil, 10).
			WillReturnError(errors.New("error"))
		m, err := s.Create(t.Context(), model.NewMetricCounter("PollCount", 10).ToRequest())
		require.Error(t, err)
		assert.Nil(t, m)
	})
	mock.ExpectClose()
}

func TestPGXStorage_Update(t *testing.T) {
	s, mock := dbMockPGXStorage(t)
	defer func() {
		assert.NoError(t, s.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	}()
	t.Run("ok", func(t *testing.T) {
		mock.ExpectQuery("UPDATE metrics SET value = $1, delta = $2 WHERE type = $3 AND id = $4 RETURNING type, id, value, delta;").
			WithArgs(nil, 10, "counter", "PollCount").
			WillReturnRows(sqlmock.NewRows([]string{"type", "id", "value", "delta"}).AddRow("counter", "PollCount", nil, 10))
		m, err := s.Update(t.Context(), model.NewMetricCounter("PollCount", 10).ToRequest())
		require.NoError(t, err)
		assert.Equal(t, m, model.NewMetricCounter("PollCount", 10))
	})
	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery("UPDATE metrics SET value = $1, delta = $2 WHERE type = $3 AND id = $4 RETURNING type, id, value, delta;").
			WithArgs(nil, 10, "counter", "PollCount").
			WillReturnError(errors.New("error"))
		m, err := s.Update(t.Context(), model.NewMetricCounter("PollCount", 10).ToRequest())
		require.Error(t, err)
		assert.Nil(t, m)
	})
	mock.ExpectClose()
}

func TestPGXStorage_CreateOrUpdateBatch(t *testing.T) {
	s, mock := dbMockPGXStorage(t)
	defer func() {
		assert.NoError(t, s.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	}()
	t.Run("ok", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4) ON CONFLICT (type, id) DO UPDATE SET value = EXCLUDED.value, delta = EXCLUDED.delta;").
			WithArgs("counter", "PollCount", nil, 10).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectQuery("SELECT type, id, value, delta FROM metrics WHERE (type=$1 AND id=$2) ORDER BY type, id;").
			WithArgs("counter", "PollCount").
			WillReturnRows(sqlmock.NewRows([]string{"type", "id", "value", "delta"}).AddRow("counter", "PollCount", nil, 10))
		ms, err := s.CreateOrUpdateBatch(t.Context(), []*model.MetricRequest{
			model.NewMetricCounter("PollCount", 10).ToRequest(),
		})
		require.NoError(t, err)
		assert.Equal(t, ms, []*model.Metric{
			model.NewMetricCounter("PollCount", 10),
		})
	})
	t.Run("error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO metrics (type, id, value, delta) VALUES ($1, $2, $3, $4) ON CONFLICT (type, id) DO UPDATE SET value = EXCLUDED.value, delta = EXCLUDED.delta;").
			WithArgs("counter", "PollCount", nil, 10).
			WillReturnError(errors.New("error"))
		mock.ExpectRollback()
		ms, err := s.CreateOrUpdateBatch(t.Context(), []*model.MetricRequest{
			model.NewMetricCounter("PollCount", 10).ToRequest(),
		})
		require.Error(t, err)
		assert.Nil(t, ms)
	})
	mock.ExpectClose()
}
