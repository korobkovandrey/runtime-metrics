package utils

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestGetRuntimeMetrics(t *testing.T) {
	metricsOk := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle",
		"HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs",
		"StackInuse", "StackSys", "Sys", "TotalAlloc"}
	gotResult := GetRuntimeMetrics()
	for _, m := range metricsOk {
		assert.Contains(t, gotResult, m)
	}
	for m := range gotResult {
		assert.Contains(t, metricsOk, m)
	}
}
