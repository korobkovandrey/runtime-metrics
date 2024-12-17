package model

type Metric struct {
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
}

func NewMetricGauge(id string, value float64) *Metric {
	return &Metric{
		Value: &value,
		ID:    id,
		MType: "gauge",
	}
}

func NewMetricCounter(id string, delta int64) *Metric {
	return &Metric{
		Delta: &delta,
		ID:    id,
		MType: "counter",
	}
}
