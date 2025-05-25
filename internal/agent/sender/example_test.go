package sender_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

// float64Ptr creates a pointer to a float64
func float64Ptr(f float64) *float64 {
	return &f
}

// int64Ptr creates a pointer to an int64
func int64Ptr(i int64) *int64 {
	return &i
}

func Example() {
	// Create a logger
	logger, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		return
	}

	// Create a mock server to receive metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a Sender configuration
	cfg := &sender.Config{
		UpdateURL:   server.URL + "/update",
		UpdatesURL:  server.URL + "/updates",
		Timeout:     5 * time.Second,
		RetryDelays: []time.Duration{1 * time.Second, 3 * time.Second},
		Key:         []byte("secret-key"),
		RateLimit:   2,
	}

	// Create a Sender
	s := sender.New(cfg, logger)
	ctx := context.Background()

	// Create a sample metric
	metric := &model.Metric{
		ID:    "testGauge",
		MType: "gauge",
		Value: float64Ptr(42.0),
	}

	// Send a single metric
	err = s.SendMetric(ctx, metric)
	if err != nil {
		fmt.Printf("Error sending metric: %v\n", err)
		return
	}

	// Send a batch of metrics
	batch := []*model.Metric{
		{ID: "testGauge2", MType: "gauge", Value: float64Ptr(43.0)},
		{ID: "testCounter", MType: "counter", Delta: int64Ptr(100)},
	}
	err = s.SendBatchMetrics(ctx, batch)
	if err != nil {
		fmt.Printf("Error sending batch metrics: %v\n", err)
		return
	}

	// Send metrics in parallel
	results := s.SendPoolMetrics(ctx, 2, batch)
	for result := range results {
		if result.Err != nil {
			fmt.Printf("Error sending metric %s: %v\n", result.Metric.ID, result.Err)
		}
	}

	fmt.Println("All metrics sent successfully")
	// Output: All metrics sent successfully
}

func ExampleSender_SendMetric() {
	// Create a logger
	logger, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		return
	}

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a Sender configuration
	cfg := &sender.Config{
		UpdateURL:   server.URL + "/update",
		Timeout:     5 * time.Second,
		RetryDelays: []time.Duration{1 * time.Second},
		Key:         []byte("secret-key"),
	}

	// Create a Sender
	s := sender.New(cfg, logger)
	ctx := context.Background()

	// Create a sample metric
	metric := &model.Metric{
		ID:    "testGauge",
		MType: "gauge",
		Value: float64Ptr(42.0),
	}

	// Send the metric
	err = s.SendMetric(ctx, metric)
	if err != nil {
		fmt.Printf("Error sending metric: %v\n", err)
		return
	}

	fmt.Printf("Sent metric: %s, Type: %s, Value: %.1f\n", metric.ID, metric.MType, *metric.Value)
	// Output: Sent metric: testGauge, Type: gauge, Value: 42.0
}

func ExampleSender_SendBatchMetrics() {
	// Create a logger
	logger, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		return
	}

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a Sender configuration
	cfg := &sender.Config{
		UpdatesURL:  server.URL + "/updates",
		Timeout:     5 * time.Second,
		RetryDelays: []time.Duration{1 * time.Second},
		Key:         []byte("secret-key"),
	}

	// Create a Sender
	s := sender.New(cfg, logger)
	ctx := context.Background()

	// Create a batch of metrics
	batch := []*model.Metric{
		{ID: "testGauge", MType: "gauge", Value: float64Ptr(42.0)},
		{ID: "testCounter", MType: "counter", Delta: int64Ptr(100)},
	}

	// Send the batch
	err = s.SendBatchMetrics(ctx, batch)
	if err != nil {
		fmt.Printf("Error sending batch: %v\n", err)
		return
	}

	fmt.Printf("Sent %d metrics\n", len(batch))
	// Output: Sent 2 metrics
}

func ExampleSender_SendPoolMetrics() {
	// Create a logger
	logger, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		return
	}

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a Sender configuration
	cfg := &sender.Config{
		UpdateURL:   server.URL + "/update",
		Timeout:     5 * time.Second,
		RetryDelays: []time.Duration{1 * time.Second},
		Key:         []byte("secret-key"),
		RateLimit:   2,
	}

	// Create a Sender
	s := sender.New(cfg, logger)
	ctx := context.Background()

	// Create a batch of metrics
	batch := []*model.Metric{
		{ID: "testGauge", MType: "gauge", Value: float64Ptr(42.0)},
		{ID: "testCounter", MType: "counter", Delta: int64Ptr(100)},
	}

	// Send metrics in parallel
	results := s.SendPoolMetrics(ctx, 2, batch)
	successCount := 0
	for result := range results {
		if result.Err == nil {
			successCount++
		}
	}

	fmt.Printf("Successfully sent %d metrics\n", successCount)
	// Output: Successfully sent 2 metrics
}
