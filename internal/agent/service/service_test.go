package service

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

var testRuntimeMetricNames = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
	"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
	"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs",
	"StackInuse", "StackSys", "Sys", "TotalAlloc"}

func TestSource_NewGaugeSource_Collect_GetDataForSend(t *testing.T) {
	s := NewGaugeSource()
	s.Collect()
	assert.Len(t, s.data, len(testRuntimeMetricNames)+1)
	assert.Contains(t, s.data, "RandomValue")
	assert.NotEmpty(t, s.data["RandomValue"])
	for _, m := range testRuntimeMetricNames {
		assert.Contains(t, s.data, m)
	}

	result := s.GetDataForSend()
	assert.Len(t, result, len(testRuntimeMetricNames)+1)
	assert.Contains(t, result, "RandomValue")
	assert.NotEqual(t, result["RandomValue"], "0")
	for _, m := range testRuntimeMetricNames {
		assert.Contains(t, result, m)
	}
}
