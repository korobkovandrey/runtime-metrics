package pgxstorage

import (
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMakeFindBatchQuery(t *testing.T) {
	tests := []struct {
		name       string
		input      []*model.MetricRequest
		wantQuery  string
		wantParams []any
	}{
		{
			name: "single metric request",
			input: []*model.MetricRequest{
				{Metric: &model.Metric{MType: "gauge", ID: "metric1"}},
			},
			wantQuery:  "SELECT type, id, value, delta FROM metrics WHERE (type=$1 AND id=$2) ORDER BY type, id;",
			wantParams: []any{"gauge", "metric1"},
		},
		{
			name: "multiple metric requests",
			input: []*model.MetricRequest{
				{Metric: &model.Metric{MType: "gauge", ID: "metric1"}},
				{Metric: &model.Metric{MType: "counter", ID: "metric2"}},
			},
			wantQuery:  "SELECT type, id, value, delta FROM metrics WHERE (type=$1 AND id=$2) OR (type=$3 AND id=$4) ORDER BY type, id;",
			wantParams: []any{"gauge", "metric1", "counter", "metric2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, params := makeFindBatchQuery(tt.input)
			assert.Equal(t, tt.wantQuery, query)
			assert.Equal(t, tt.wantParams, params)
		})
	}
}

func Test_prepareStatements(t *testing.T) {
	s, mock := dbMockPGXStorage(t)
	defer func() {
		assert.NoError(t, s.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	}()
	mock.ExpectClose()
}
