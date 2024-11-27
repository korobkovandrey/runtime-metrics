package adapter

import (
	"strconv"
)

type Counter struct {
	Repository
}

func (m Counter) Update(name string, value string) error {
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	m.IncrCounter(name, number)
	return nil
}

func NewCounter(storage Repository) *Counter {
	return &Counter{storage}
}
