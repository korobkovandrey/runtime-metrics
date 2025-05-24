package service

import (
	"context"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenGopsutilMetrics(t *testing.T) {
	ch := genGopsutilMetrics(context.Background())
	var metrics []*model.Metric
	for res := range ch {
		require.NoError(t, res.err)
		metrics = append(metrics, res.m)
	}

	assert.GreaterOrEqual(t, len(metrics), 2)
	hasTotalMemory := false
	hasFreeMemory := false
	for _, m := range metrics {
		if m.ID == "TotalMemory" {
			hasTotalMemory = true
		}
		if m.ID == "FreeMemory" {
			hasFreeMemory = true
		}
	}
	assert.True(t, hasTotalMemory)
	assert.True(t, hasFreeMemory)
}

func TestGenPullMetrics(t *testing.T) {
	ch := genPullMetrics()
	var metrics []*model.Metric
	for m := range ch {
		metrics = append(metrics, m)
	}

	assert.NotEmpty(t, metrics)
	hasRandomValue := false
	for _, m := range metrics {
		if m.ID == "RandomValue" {
			hasRandomValue = true
			assert.GreaterOrEqual(t, *m.Value, 0.0)
			assert.Less(t, *m.Value, 1.0)
		}
	}
	assert.True(t, hasRandomValue)
}

func TestGetRuntimeMetrics(t *testing.T) {
	metrics := getRuntimeMetrics()

	expectedKeys := []string{"Alloc", "HeapAlloc", "TotalAlloc", "Frees"}
	for _, key := range expectedKeys {
		_, ok := metrics[key]
		assert.True(t, ok, "Expected key %s in metrics", key)
	}

	for key, value := range metrics {
		assert.GreaterOrEqual(t, value, 0.0, "Expected non-negative value for %s", key)
	}
}
