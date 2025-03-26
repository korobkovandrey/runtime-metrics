package service

import (
	"context"
	"math/rand/v2"
	"runtime"
	"strconv"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type genResult struct {
	m   *model.Metric
	err error
}

func genGopsutilMetrics(ctx context.Context) <-chan genResult {
	out := make(chan genResult)
	go func() {
		defer close(out)
		v, err := mem.VirtualMemoryWithContext(ctx)
		if err != nil {
			out <- genResult{err: err}
			return
		}
		out <- genResult{m: model.NewMetricGauge("TotalMemory", float64(v.Total))}
		out <- genResult{m: model.NewMetricGauge("FreeMemory", float64(v.Free))}
		pcts, err := cpu.PercentWithContext(ctx, 0, false)
		if err != nil {
			out <- genResult{err: err}
			return
		}
		for i, pct := range pcts {
			out <- genResult{
				m: model.NewMetricGauge("CPUutilization"+strconv.Itoa(i+1), pct),
			}
		}
	}()
	return out
}

func genPullMetrics() <-chan *model.Metric {
	out := make(chan *model.Metric)
	go func() {
		defer close(out)
		for k, v := range getRuntimeMetrics() {
			out <- model.NewMetricGauge(k, v)
		}
		out <- model.NewMetricGauge("RandomValue", rand.Float64()) //nolint:gosec // ignore
	}()
	return out
}

func getRuntimeMetrics() (result map[string]float64) {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return map[string]float64{
		"Alloc":         float64(memStats.Alloc),
		"BuckHashSys":   float64(memStats.BuckHashSys),
		"Frees":         float64(memStats.Frees),
		"GCCPUFraction": memStats.GCCPUFraction,
		"GCSys":         float64(memStats.GCSys),
		"HeapAlloc":     float64(memStats.HeapAlloc),
		"HeapIdle":      float64(memStats.HeapIdle),
		"HeapInuse":     float64(memStats.HeapInuse),
		"HeapObjects":   float64(memStats.HeapObjects),
		"HeapReleased":  float64(memStats.HeapReleased),
		"HeapSys":       float64(memStats.HeapSys),
		"LastGC":        float64(memStats.LastGC),
		"Lookups":       float64(memStats.Lookups),
		"MCacheInuse":   float64(memStats.MCacheInuse),
		"MCacheSys":     float64(memStats.MCacheSys),
		"MSpanInuse":    float64(memStats.MSpanInuse),
		"MSpanSys":      float64(memStats.MSpanSys),
		"Mallocs":       float64(memStats.Mallocs),
		"NextGC":        float64(memStats.NextGC),
		"NumForcedGC":   float64(memStats.NumForcedGC),
		"NumGC":         float64(memStats.NumGC),
		"OtherSys":      float64(memStats.OtherSys),
		"PauseTotalNs":  float64(memStats.PauseTotalNs),
		"StackInuse":    float64(memStats.StackInuse),
		"StackSys":      float64(memStats.StackSys),
		"Sys":           float64(memStats.Sys),
		"TotalAlloc":    float64(memStats.TotalAlloc),
	}
}
