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

//nolint:wrapcheck // ignore
func (m *Metric) ScanRow(row pgx.Row) error {
	return row.Scan(
		&m.MType,
		&m.ID,
		&m.Value,
		&m.Delta,
	)
}

func NewMetricGauge(id string, value float64) *Metric {
	return &Metric{
		Value: &value,
		MType: TypeGauge,
		ID:    id,
	}
}

func NewMetricCounter(id string, delta int64) *Metric {
	return &Metric{
		Delta: &delta,
		MType: TypeCounter,
		ID:    id,
	}
}

type MetricRequest struct {
	*Metric
}

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

func (mr *MetricRequest) ValidateType() error {
	if mr.MType != TypeGauge && mr.MType != TypeCounter {
		return ErrTypeIsNotValid
	}
	return nil
}

func NewMetricRequest(t, id, value string) (*MetricRequest, error) {
	var m *Metric
	switch t {
	case TypeGauge:
		number, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("NewMetricRequest %w: %w", ErrValueIsNotValid, err)
		}
		m = NewMetricGauge(id, number)
	case TypeCounter:
		number, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("NewMetricRequest %w: %w", ErrValueIsNotValid, err)
		}
		m = NewMetricCounter(id, number)
	default:
		return nil, ErrTypeIsNotValid
	}
	return &MetricRequest{m}, nil
}

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
