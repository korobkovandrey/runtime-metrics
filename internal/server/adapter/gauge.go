package adapter

import (
	"fmt"
	"strconv"
)

type Gauge struct {
	Repository
	key string
}

func NewGauge(storage Repository, key string) *Gauge {
	return &Gauge{storage, key}
}

func (a Gauge) Update(name string, value string) error {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("gauge: %w", err)
	}
	a.Set(a.key, name, number)
	return nil
}

func (a Gauge) GetStorageValue(name string) (any, bool) {
	return a.Get(a.key, name)
}

func (a Gauge) Names() []string {
	return a.Keys(a.key)
}
