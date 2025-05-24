package service

import (
	"context"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

var testMetricNames = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
	"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
	"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs",
	"StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue", "TotalMemory", "FreeMemory"}

func TestSource_NewGaugeSource_Collect_GetDataForSend(t *testing.T) {
	s := NewSource()
	err := s.Collect(context.TODO())
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(s.data), len(testMetricNames)+1)
	gaugesMap := make(map[string]*model.Metric)
	for _, m := range s.data {
		switch m.MType {
		case model.TypeGauge:
			gaugesMap[m.ID] = m
		case model.TypeCounter:
			t.Errorf("metric type should be gauge: %s", m.MType)
		default:
			t.Errorf("unexpected metric type: %s", m.MType)
		}
	}
	require.Contains(t, gaugesMap, "RandomValue")
	assert.NotEmpty(t, gaugesMap["RandomValue"])
	for _, name := range testMetricNames {
		assert.Contains(t, gaugesMap, name)
	}

	data, delta := s.Get()
	assert.Equal(t, int64(1), delta)
	assert.GreaterOrEqual(t, len(data), len(testMetricNames)+2)
	gaugesMap = make(map[string]*model.Metric)
	counterMap := make(map[string]*model.Metric)
	for _, m := range data {
		switch m.MType {
		case model.TypeGauge:
			gaugesMap[m.ID] = m
		case model.TypeCounter:
			counterMap[m.ID] = m
		default:
			t.Errorf("unexpected metric type: %s", m.MType)
		}
	}
	require.Contains(t, gaugesMap, "RandomValue")
	assert.NotEmpty(t, gaugesMap["RandomValue"])
	require.Contains(t, counterMap, "PollCount")
	assert.Equal(t, int64(1), *counterMap["PollCount"].Delta)
	for _, m := range testMetricNames {
		assert.Contains(t, gaugesMap, m)
	}
}

func TestSource_Collect(t *testing.T) {
	t.Run("Successful collection", func(t *testing.T) {
		s := NewSource()
		err := s.Collect(context.TODO())
		require.NoError(t, err)
		assert.NotEmpty(t, s.data)
		assert.Equal(t, int64(1), *s.pollCount.Delta)
	})
}

func TestSource_Get(t *testing.T) {
	s := NewSource()
	s.data = []*model.Metric{
		model.NewMetricGauge("TestMetric", 42.0),
	}
	*s.pollCount.Delta = 5

	data, delta := s.Get()

	t.Run("Correct length and content", func(t *testing.T) {
		assert.Len(t, data, 2)
		assert.Equal(t, "TestMetric", data[0].ID)
		assert.Equal(t, 42.0, *data[0].Value)
		assert.Equal(t, "PollCount", data[1].ID)
		assert.Equal(t, int64(5), *data[1].Delta)
		assert.Equal(t, int64(5), delta)
	})

	t.Run("Deep copy", func(t *testing.T) {
		*data[0].Value = 100.0
		assert.Equal(t, 42.0, *s.data[0].Value)
	})
}

func TestSource_Commit(t *testing.T) {
	s := NewSource()
	*s.pollCount.Delta = 10

	t.Run("Reduce delta", func(t *testing.T) {
		s.Commit(3)
		assert.Equal(t, int64(7), *s.pollCount.Delta)
	})

	t.Run("Negative delta", func(t *testing.T) {
		s.Commit(10)
		assert.Equal(t, int64(-3), *s.pollCount.Delta)
	})
}
