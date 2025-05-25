package service_test

import (
	"context"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/service"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

func Example() {
	// Create a new Source
	source := service.NewSource()
	ctx := context.Background()

	// Collect metrics
	err := source.Collect(ctx)
	if err != nil {
		fmt.Printf("Error collecting metrics: %v\n", err)
		return
	}

	// Get metrics and poll count
	metrics, delta := source.Get()

	// Commit metrics
	source.Commit(delta)

	// Print a subset of collected metrics (example metrics, actual count may vary)
	p := map[string]*model.Metric{}
	for _, m := range metrics {
		if m.ID == "Alloc" || m.ID == "TotalMemory" || m.ID == "PollCount" {
			p[m.ID] = m
		}
	}
	if _, ok := p["Alloc"]; ok {
		fmt.Printf("Metric: %s, Type: %s, Value not nil: %v\n", p["Alloc"].ID, p["Alloc"].MType, p["Alloc"].Value != nil)
	}
	if _, ok := p["TotalMemory"]; ok {
		fmt.Printf("Metric: %s, Type: %s, Value not nil: %v\n", p["TotalMemory"].ID, p["TotalMemory"].MType, p["Alloc"].Value != nil)
	}
	if _, ok := p["PollCount"]; ok {
		fmt.Printf("Metric: %s, Type: %s, Delta: %d\n", p["PollCount"].ID, p["PollCount"].MType, *p["PollCount"].Delta)
	}

	// Get metrics again to check poll count after commit
	_, newDelta := source.Get()
	fmt.Printf("Poll count after commit: %d\n", newDelta)
	// Output:
	// Metric: Alloc, Type: gauge, Value not nil: true
	// Metric: TotalMemory, Type: gauge, Value not nil: true
	// Metric: PollCount, Type: counter, Delta: 1
	// Poll count after commit: 0
}

func ExampleSource_Collect() {
	// Create a new Source
	source := service.NewSource()
	ctx := context.Background()

	// Collect metrics
	err := source.Collect(ctx)
	if err != nil {
		fmt.Printf("Error collecting metrics: %v\n", err)
		return
	}

	// Get metrics to verify collection
	metrics, _ := source.Get()
	fmt.Printf("Collected %d metrics\n", len(metrics))
	// Output: Collected 32 metrics
}

func ExampleSource_Get() {
	// Create a new Source
	source := service.NewSource()
	ctx := context.Background()

	// Collect metrics
	err := source.Collect(ctx)
	if err != nil {
		fmt.Printf("Error collecting metrics: %v\n", err)
		return
	}

	// Get metrics and poll count
	metrics, delta := source.Get()
	p := map[string]*model.Metric{}
	for _, m := range metrics {
		if m.ID == "HeapAlloc" || m.ID == "FreeMemory" || m.ID == "PollCount" {
			p[m.ID] = m
		}
	}
	if _, ok := p["HeapAlloc"]; ok {
		fmt.Printf("Metric: %s, Type: %s, Value not nil: %v\n", p["HeapAlloc"].ID, p["HeapAlloc"].MType, p["HeapAlloc"].Value != nil)
	}
	if _, ok := p["FreeMemory"]; ok {
		fmt.Printf("Metric: %s, Type: %s, Value not nil: %v\n", p["FreeMemory"].ID, p["FreeMemory"].MType, p["FreeMemory"].Value != nil)
	}
	if _, ok := p["PollCount"]; ok {
		fmt.Printf("Metric: %s, Type: %s, Delta: %d\n", p["PollCount"].ID, p["PollCount"].MType, *p["PollCount"].Delta)
	}
	fmt.Printf("Poll count: %d\n", delta)
	// Output:
	// Metric: HeapAlloc, Type: gauge, Value not nil: true
	// Metric: FreeMemory, Type: gauge, Value not nil: true
	// Metric: PollCount, Type: counter, Delta: 1
	// Poll count: 1
}

func ExampleSource_Commit() {
	// Create a new Source
	source := service.NewSource()
	ctx := context.Background()

	// Collect metrics
	err := source.Collect(ctx)
	if err != nil {
		fmt.Printf("Error collecting metrics: %v\n", err)
		return
	}

	// Get poll count
	_, delta := source.Get()

	// Commit metrics
	source.Commit(delta)

	// Verify poll count after commit
	_, newDelta := source.Get()
	fmt.Printf("Poll count after commit: %d\n", newDelta)
	// Output: Poll count after commit: 0
}
