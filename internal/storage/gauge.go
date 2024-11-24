package storage

import (
	"strconv"
)

type Gauge struct {
	Storage
}

func (m Gauge) Handler(name string, value string) error {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	m.SetGauge(name, number)
	return nil
}
