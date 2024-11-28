package adapter

import (
	"fmt"
	"strconv"
)

type Gauge struct {
	Repository
	key string
}

func (a Gauge) Update(name string, value string) error {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf(`gauge: %w`, err)
	}
	a.Set(a.key, name, number)
	return nil
}

func NewGauge(storage Repository, key string) *Gauge {
	return &Gauge{storage, key}
}
