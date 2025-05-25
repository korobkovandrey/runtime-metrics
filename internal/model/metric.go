// Package model contains structures for work with metrics.
package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/jackc/pgx/v5"
)

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

// Metric - metric structure
type Metric struct {
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	ID    string   `json:"id"`              // имя метрики
}

// Clone returns a copy of the metric
func (m *Metric) Clone() *Metric {
	metric := &Metric{
		MType: m.MType,
		ID:    m.ID,
	}
	if m.Value != nil {
		metric.Value = new(float64)
		*metric.Value = *m.Value
	}
	if m.Delta != nil {
		metric.Delta = new(int64)
		*metric.Delta = *m.Delta
	}
	return metric
}

// AnyValue returns the value of the metric as any
func (m *Metric) AnyValue() any {
	if m.MType == TypeCounter {
		if m.Delta == nil {
			return nil
		}
		return *m.Delta
	}
	if m.Value == nil {
		return nil
	}
	return *m.Value
}

// ScanRow scans the row into the metric
func (m *Metric) ScanRow(row pgx.Row) error {
	return row.Scan(
		&m.MType,
		&m.ID,
		&m.Value,
		&m.Delta,
	)
}

// NewMetricGauge returns a new gauge metric
func NewMetricGauge(id string, value float64) *Metric {
	return &Metric{
		Value: &value,
		MType: TypeGauge,
		ID:    id,
	}
}

// NewMetricCounter returns a new counter metric
func NewMetricCounter(id string, delta int64) *Metric {
	return &Metric{
		Delta: &delta,
		MType: TypeCounter,
		ID:    id,
	}
}

// MetricRequest - metric request structure
type MetricRequest struct {
	*Metric
}

// RequiredValue returns an error if the value is not set
func (mr *MetricRequest) RequiredValue() error {
	switch mr.MType {
	case TypeGauge:
		if mr.Value == nil {
			return ErrValueIsNotValid
		}
	case TypeCounter:
		if mr.Delta == nil {
			return ErrValueIsNotValid
		}
	default:
		return ErrTypeIsNotValid
	}
	return nil
}

// ValidateType returns an error if the type is not valid
func (mr *MetricRequest) ValidateType() error {
	if mr.MType != TypeGauge && mr.MType != TypeCounter {
		return ErrTypeIsNotValid
	}
	return nil
}

// NewMetricRequest returns a new metric request
func NewMetricRequest(t, id, value string) (*MetricRequest, error) {
	var m *Metric
	switch t {
	case TypeGauge:
		number, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrValueIsNotValid, err)
		}
		m = NewMetricGauge(id, number)
	case TypeCounter:
		number, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrValueIsNotValid, err)
		}
		m = NewMetricCounter(id, number)
	default:
		return nil, ErrTypeIsNotValid
	}
	return &MetricRequest{m}, nil
}

// UnmarshalMetricRequestFromReader unmarshals the metric request from the reader
func UnmarshalMetricRequestFromReader(r io.Reader) (*MetricRequest, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	if len(body) == 0 {
		return nil, errors.New("empty body")
	}
	var metric *MetricRequest
	err = json.Unmarshal(body, &metric)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %w", err)
	}
	if metric == nil || metric.Metric == nil {
		return nil, ErrMetricNotFound
	}
	return metric, nil
}

// UnmarshalMetricsRequestFromReader unmarshals the metric request from the reader
func UnmarshalMetricsRequestFromReader(r io.Reader) ([]*MetricRequest, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	if len(body) == 0 {
		return nil, errors.New("empty body")
	}
	var metrics []*MetricRequest
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %w", err)
	}
	if len(metrics) == 0 {
		return nil, errors.New("empty body")
	}
	return metrics, nil
}

// ValidateMetricsRequest returns an error if the metric request is not valid
func ValidateMetricsRequest(metrics []*MetricRequest) error {
	for _, m := range metrics {
		if err := m.RequiredValue(); err != nil {
			return err
		}
	}
	return nil
}
