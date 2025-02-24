package model

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

type Metric struct {
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	ID    string   `json:"id"`              // имя метрики
}

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

func (m *Metric) AnyValue() any {
	if m.MType == TypeCounter {
		return *m.Delta
	}
	return *m.Value
}

func NewMetricGauge(id string, value float64) Metric {
	return Metric{
		Value: &value,
		MType: TypeGauge,
		ID:    id,
	}
}

func NewMetricCounter(id string, delta int64) Metric {
	return Metric{
		Delta: &delta,
		MType: TypeCounter,
		ID:    id,
	}
}

type MetricRequest struct {
	Metric
}

func NewMetricRequest(t string, name string, value string) (*MetricRequest, error) {
	var m Metric
	switch t {
	case TypeGauge:
		number, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("NewMetricRequest %w: %w", ErrValueIsNotValid, err)
		}
		m = NewMetricGauge(name, number)
	case TypeCounter:
		number, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("NewMetricRequest %w: %w", ErrValueIsNotValid, err)
		}
		m = NewMetricCounter(name, number)
	default:
		return nil, ErrTypeIsNotValid
	}
	return &MetricRequest{m}, nil
}

func NewMetricRequestFromReader(r io.Reader) (*MetricRequest, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("readMetricFromRequest %w: %w", ErrMetricNotFound, err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("readMetricFromRequest %w", ErrMetricNotFound)
	}
	var metric *MetricRequest
	err = json.Unmarshal(body, &metric)
	if err != nil {
		return nil, fmt.Errorf("readMetricFromRequest %w: %w", ErrMetricNotFound, err)
	}
	return metric, nil
}
