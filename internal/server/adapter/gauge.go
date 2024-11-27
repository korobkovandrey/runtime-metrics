package adapter

import (
	"strconv"
)

type Gauge struct {
	Repository
}

func (m Gauge) Update(name string, value string) error {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	m.SetGauge(name, number)
	return nil
}

func NewGauge(storage Repository) *Gauge {
	return &Gauge{storage}
}
