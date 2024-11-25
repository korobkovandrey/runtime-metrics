package repository

import (
	"github.com/korobkovandrey/runtime-metrics/internal/storage"
	"strconv"
)

type Counter struct {
	storage.Storage
}

func (m Counter) GetType() Type {
	return counterType
}

func (m Counter) Update(name string, value string) error {
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	m.IncrCounter(name, number)
	return nil
}

func NewCounter(storage storage.Storage) Counter {
	return Counter{storage}
}
