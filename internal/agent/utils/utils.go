package utils

import (
	"runtime"
)

func GetRuntimeMetricsFloat64() (result map[string]float64) {
	result = map[string]float64{}
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	result["Alloc"] = float64(memStats.Alloc)
	result["BuckHashSys"] = float64(memStats.BuckHashSys)
	result["Frees"] = float64(memStats.Frees)
	result["GCCPUFraction"] = memStats.GCCPUFraction
	result["GCSys"] = float64(memStats.GCSys)
	result["HeapAlloc"] = float64(memStats.HeapAlloc)
	result["HeapIdle"] = float64(memStats.HeapIdle)
	result["HeapInuse"] = float64(memStats.HeapInuse)
	result["HeapObjects"] = float64(memStats.HeapObjects)
	result["HeapReleased"] = float64(memStats.HeapReleased)
	result["HeapSys"] = float64(memStats.HeapSys)
	result["LastGC"] = float64(memStats.LastGC)
	result["Lookups"] = float64(memStats.Lookups)
	result["MCacheInuse"] = float64(memStats.MCacheInuse)
	result["MCacheSys"] = float64(memStats.MCacheSys)
	result["MSpanInuse"] = float64(memStats.MSpanInuse)
	result["MSpanSys"] = float64(memStats.MSpanSys)
	result["Mallocs"] = float64(memStats.Mallocs)
	result["NextGC"] = float64(memStats.NextGC)
	result["NumForcedGC"] = float64(memStats.NumForcedGC)
	result["NumGC"] = float64(memStats.NumGC)
	result["OtherSys"] = float64(memStats.OtherSys)
	result["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	result["StackInuse"] = float64(memStats.StackInuse)
	result["StackSys"] = float64(memStats.StackSys)
	result["Sys"] = float64(memStats.Sys)
	result["TotalAlloc"] = float64(memStats.TotalAlloc)
	return
}
