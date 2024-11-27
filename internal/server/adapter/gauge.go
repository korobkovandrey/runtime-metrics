package adapter

import (
	"fmt"
	"strconv"
)

type Gauge struct {
	Repository
}

func (m Gauge) Update(name string, value string) error {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf(`gauge: %w`, err)
	}
	m.SetGauge(name, number)
	return nil
}

func NewGauge(storage Repository) *Gauge {
	return &Gauge{storage}
}
