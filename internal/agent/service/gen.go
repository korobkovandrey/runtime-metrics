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
		memStats := &runtime.MemStats{}
		runtime.ReadMemStats(memStats)
		out <- model.NewMetricGauge("Alloc", float64(memStats.Alloc))
		out <- model.NewMetricGauge("BuckHashSys", float64(memStats.BuckHashSys))
		out <- model.NewMetricGauge("Frees", float64(memStats.Frees))
		out <- model.NewMetricGauge("GCCPUFraction", memStats.GCCPUFraction)
		out <- model.NewMetricGauge("GCSys", float64(memStats.GCSys))
		out <- model.NewMetricGauge("HeapAlloc", float64(memStats.HeapAlloc))
		out <- model.NewMetricGauge("HeapIdle", float64(memStats.HeapIdle))
		out <- model.NewMetricGauge("HeapInuse", float64(memStats.HeapInuse))
		out <- model.NewMetricGauge("HeapObjects", float64(memStats.HeapObjects))
		out <- model.NewMetricGauge("HeapReleased", float64(memStats.HeapReleased))
		out <- model.NewMetricGauge("HeapSys", float64(memStats.HeapSys))
		out <- model.NewMetricGauge("LastGC", float64(memStats.LastGC))
		out <- model.NewMetricGauge("Lookups", float64(memStats.Lookups))
		out <- model.NewMetricGauge("MCacheInuse", float64(memStats.MCacheInuse))
		out <- model.NewMetricGauge("MCacheSys", float64(memStats.MCacheSys))
		out <- model.NewMetricGauge("MSpanInuse", float64(memStats.MSpanInuse))
		out <- model.NewMetricGauge("MSpanSys", float64(memStats.MSpanSys))
		out <- model.NewMetricGauge("Mallocs", float64(memStats.Mallocs))
		out <- model.NewMetricGauge("NextGC", float64(memStats.NextGC))
		out <- model.NewMetricGauge("NumForcedGC", float64(memStats.NumForcedGC))
		out <- model.NewMetricGauge("NumGC", float64(memStats.NumGC))
		out <- model.NewMetricGauge("OtherSys", float64(memStats.OtherSys))
		out <- model.NewMetricGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
		out <- model.NewMetricGauge("StackInuse", float64(memStats.StackInuse))
		out <- model.NewMetricGauge("StackSys", float64(memStats.StackSys))
		out <- model.NewMetricGauge("Sys", float64(memStats.Sys))
		out <- model.NewMetricGauge("TotalAlloc", float64(memStats.TotalAlloc))
		out <- model.NewMetricGauge("RandomValue", rand.Float64()) //nolint:gosec // ignore
	}()
	return out
}
