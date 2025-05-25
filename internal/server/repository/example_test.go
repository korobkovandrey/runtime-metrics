package repository_test

import (
	"context"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
)

func Example() {
	// Create a new MemStorage
	storage := repository.NewMemStorage()
	ctx := context.Background()

	// Create a new metric
	mr := &model.MetricRequest{
		Metric: &model.Metric{
			ID:    "testGauge",
			MType: "gauge",
			Value: float64Ptr(42.0),
		},
	}
	metric, err := storage.Create(ctx, mr)
	if err != nil {
		fmt.Printf("Error creating metric: %v\n", err)
		return
	}

	// Update the metric
	mr.Value = float64Ptr(43.0)
	updated, err := storage.Update(ctx, mr)
	if err != nil {
		fmt.Printf("Error updating metric: %v\n", err)
		return
	}

	// Find the metric
	found, err := storage.Find(ctx, mr)
	if err != nil {
		fmt.Printf("Error finding metric: %v\n", err)
		return
	}

	fmt.Printf("Created metric: %s, Value: %.1f\n", metric.ID, *metric.Value)
	fmt.Printf("Updated metric: %s, Value: %.1f\n", updated.ID, *updated.Value)
	fmt.Printf("Found metric: %s, Value: %.1f\n", found.ID, *found.Value)
	// Output:
	// Created metric: testGauge, Value: 42.0
	// Updated metric: testGauge, Value: 43.0
	// Found metric: testGauge, Value: 43.0
}

func ExampleMemStorage_Create() {
	storage := repository.NewMemStorage()
	ctx := context.Background()

	mr := &model.MetricRequest{
		Metric: &model.Metric{
			ID:    "testCounter",
			MType: "counter",
			Delta: int64Ptr(100),
		},
	}
	metric, err := storage.Create(ctx, mr)
	if err != nil {
		fmt.Printf("Error creating metric: %v\n", err)
		return
	}

	fmt.Printf("Created metric: %s, Type: %s, Delta: %d\n", metric.ID, metric.MType, *metric.Delta)
	// Output: Created metric: testCounter, Type: counter, Delta: 100
}

func ExampleMemStorage_Update() {
	storage := repository.NewMemStorage()
	ctx := context.Background()

	// First, create a metric
	mr := &model.MetricRequest{
		Metric: &model.Metric{
			ID:    "testGauge",
			MType: "gauge",
			Value: float64Ptr(42.0),
		},
	}
	_, err := storage.Create(ctx, mr)
	if err != nil {
		fmt.Printf("Error creating metric: %v\n", err)
		return
	}

	// Update the metric
	mr.Value = float64Ptr(50.0)
	updated, err := storage.Update(ctx, mr)
	if err != nil {
		fmt.Printf("Error updating metric: %v\n", err)
		return
	}

	fmt.Printf("Updated metric: %s, Value: %.1f\n", updated.ID, *updated.Value)
	// Output: Updated metric: testGauge, Value: 50.0
}

func ExampleMemStorage_CreateOrUpdateBatch() {
	storage := repository.NewMemStorage()
	ctx := context.Background()

	mrs := []*model.MetricRequest{
		{
			Metric: &model.Metric{
				ID:    "testGauge",
				MType: "gauge",
				Value: float64Ptr(42.0),
			},
		},
		{
			Metric: &model.Metric{
				ID:    "testCounter",
				MType: "counter",
				Delta: int64Ptr(100),
			},
		},
	}

	metrics, err := storage.CreateOrUpdateBatch(ctx, mrs)
	if err != nil {
		fmt.Printf("Error in batch operation: %v\n", err)
		return
	}

	for _, m := range metrics {
		if m.MType == "gauge" {
			fmt.Printf("Metric: %s, Type: %s, Value: %.1f\n", m.ID, m.MType, *m.Value)
		} else {
			fmt.Printf("Metric: %s, Type: %s, Delta: %d\n", m.ID, m.MType, *m.Delta)
		}
	}
	// Output:
	// Metric: testGauge, Type: gauge, Value: 42.0
	// Metric: testCounter, Type: counter, Delta: 100
}

func ExampleMemStorage_Find() {
	storage := repository.NewMemStorage()
	ctx := context.Background()

	// Create a metric
	mr := &model.MetricRequest{
		Metric: &model.Metric{
			ID:    "testGauge",
			MType: "gauge",
			Value: float64Ptr(42.0),
		},
	}
	_, err := storage.Create(ctx, mr)
	if err != nil {
		fmt.Printf("Error creating metric: %v\n", err)
		return
	}

	// Find the metric
	found, err := storage.Find(ctx, mr)
	if err != nil {
		fmt.Printf("Error finding metric: %v\n", err)
		return
	}

	fmt.Printf("Found metric: %s, Value: %.1f\n", found.ID, *found.Value)
	// Output: Found metric: testGauge, Value: 42.0
}

func ExampleMemStorage_FindAll() {
	storage := repository.NewMemStorage()
	ctx := context.Background()

	// Create some metrics
	mrs := []*model.MetricRequest{
		{
			Metric: &model.Metric{
				ID:    "testGauge",
				MType: "gauge",
				Value: float64Ptr(42.0),
			},
		},
		{
			Metric: &model.Metric{
				ID:    "testCounter",
				MType: "counter",
				Delta: int64Ptr(100),
			},
		},
	}
	_, err := storage.CreateOrUpdateBatch(ctx, mrs)
	if err != nil {
		fmt.Printf("Error creating metrics: %v\n", err)
		return
	}

	// Find all metrics
	metrics, err := storage.FindAll(ctx)
	if err != nil {
		fmt.Printf("Error finding all metrics: %v\n", err)
		return
	}

	for _, m := range metrics {
		if m.MType == "gauge" {
			fmt.Printf("Metric: %s, Type: %s, Value: %.1f\n", m.ID, m.MType, *m.Value)
		} else {
			fmt.Printf("Metric: %s, Type: %s, Delta: %d\n", m.ID, m.MType, *m.Delta)
		}
	}
	// Output:
	// Metric: testGauge, Type: gauge, Value: 42.0
	// Metric: testCounter, Type: counter, Delta: 100
}

// Helper functions to create pointers for test data
func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
