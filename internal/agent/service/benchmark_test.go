package service

import (
	"fmt"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

// BenchmarkSource_Collect измеряет производительность метода Collect.
func BenchmarkSource_Collect(b *testing.B) {
	s := NewSource()
	ctx := b.Context()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := s.Collect(ctx)
		assert.NoError(b, err, "Collect should succeed")
	}
}

// BenchmarkSource_Get измеряет производительность метода Get.
func BenchmarkSource_Get(b *testing.B) {
	s := NewSource()
	// Подготовка данных
	s.data = make([]*model.Metric, 100)
	for i := 0; i < 100; i++ {
		s.data[i] = model.NewMetricGauge(fmt.Sprintf("Metric%d", i), float64(i))
	}
	*s.pollCount.Delta = 5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, delta := s.Get()
		assert.Len(b, data, 101, "Expected 101 metrics (100 + pollCount)")
		assert.Equal(b, int64(5), delta, "Expected delta 5")
	}
}

// BenchmarkSource_Commit измеряет производительность метода Commit.
func BenchmarkSource_Commit(b *testing.B) {
	s := NewSource()
	*s.pollCount.Delta = 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Commit(1)
	}
}

// BenchmarkGenGopsutilMetrics измеряет производительность функции genGopsutilMetrics.
func BenchmarkGenGopsutilMetrics(b *testing.B) {
	ctx := b.Context()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := genGopsutilMetrics(ctx)
		for res := range ch {
			assert.NoError(b, res.err, "Expected no error from genGopsutilMetrics")
		}
	}
}

// BenchmarkGenPullMetrics измеряет производительность функции genPullMetrics.
func BenchmarkGenPullMetrics(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := genPullMetrics()
		for range ch {
			// Просто читаем канал до закрытия
		}
	}
}

// BenchmarkGetRuntimeMetrics измеряет производительность функции getRuntimeMetrics.
func BenchmarkGetRuntimeMetrics(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics := getRuntimeMetrics()
		assert.NotEmpty(b, metrics, "Expected non-empty metrics map")
	}
}
