package adapter

import (
	"fmt"
	"strconv"
)

type Counter struct {
	Repository
}

func (m Counter) Update(name string, value string) error {
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf(`counter: %w`, err)
	}
	m.IncrCounter(name, number)
	return nil
}

func NewCounter(storage Repository) *Counter {
	return &Counter{storage}
}
