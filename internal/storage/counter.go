package storage

import (
	"strconv"
)

type Counter struct {
	Storage
}

func (m Counter) Handler(name string, value string) error {
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	m.IncrCounter(name, number)
	return nil
}
