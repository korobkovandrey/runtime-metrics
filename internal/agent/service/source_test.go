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
	require.Contains(t, counterMap, "PoolCount")
	assert.Equal(t, int64(1), *counterMap["PoolCount"].Delta)
	for _, m := range testMetricNames {
		assert.Contains(t, gaugesMap, m)
	}
}
