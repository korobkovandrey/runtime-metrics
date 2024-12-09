package utils

import (
	"runtime"
	"strconv"
)

func GetRuntimeMetrics() (result map[string]string) {
	result = map[string]string{}
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	result["Alloc"] = strconv.Itoa(int(memStats.Alloc))
	result["BuckHashSys"] = strconv.Itoa(int(memStats.BuckHashSys))
	result["Frees"] = strconv.Itoa(int(memStats.Frees))
	result["GCCPUFraction"] = strconv.FormatFloat(memStats.GCCPUFraction, 'g', -1, 64)
	result["GCSys"] = strconv.Itoa(int(memStats.GCSys))
	result["HeapAlloc"] = strconv.Itoa(int(memStats.HeapAlloc))
	result["HeapIdle"] = strconv.Itoa(int(memStats.HeapIdle))
	result["HeapInuse"] = strconv.Itoa(int(memStats.HeapInuse))
	result["HeapObjects"] = strconv.Itoa(int(memStats.HeapObjects))
	result["HeapReleased"] = strconv.Itoa(int(memStats.HeapReleased))
	result["HeapSys"] = strconv.Itoa(int(memStats.HeapSys))
	result["LastGC"] = strconv.Itoa(int(memStats.LastGC))
	result["Lookups"] = strconv.Itoa(int(memStats.Lookups))
	result["MCacheInuse"] = strconv.Itoa(int(memStats.MCacheInuse))
	result["MCacheSys"] = strconv.Itoa(int(memStats.MCacheSys))
	result["MSpanInuse"] = strconv.Itoa(int(memStats.MSpanInuse))
	result["MSpanSys"] = strconv.Itoa(int(memStats.MSpanSys))
	result["Mallocs"] = strconv.Itoa(int(memStats.Mallocs))
	result["NextGC"] = strconv.Itoa(int(memStats.NextGC))
	result["NumForcedGC"] = strconv.Itoa(int(memStats.NumForcedGC))
	result["NumGC"] = strconv.Itoa(int(memStats.NumGC))
	result["OtherSys"] = strconv.Itoa(int(memStats.OtherSys))
	result["PauseTotalNs"] = strconv.Itoa(int(memStats.PauseTotalNs))
	result["StackInuse"] = strconv.Itoa(int(memStats.StackInuse))
	result["StackSys"] = strconv.Itoa(int(memStats.StackSys))
	result["Sys"] = strconv.Itoa(int(memStats.Sys))
	result["TotalAlloc"] = strconv.Itoa(int(memStats.TotalAlloc))
	return
}
