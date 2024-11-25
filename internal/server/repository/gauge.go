package repository

import (
	"github.com/korobkovandrey/runtime-metrics/internal/storage"
	"strconv"
)

type Gauge struct {
	storage.Storage
}

func (m Gauge) GetType() Type {
	return gaugeType
}

func (m Gauge) Update(name string, value string) error {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	m.SetGauge(name, number)
	return nil
}

func NewGauge(storage storage.Storage) Gauge {
	return Gauge{storage}
}
