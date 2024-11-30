package adapter

import (
	"fmt"
	"strconv"
)

type Counter struct {
	Repository
	key string
}

func NewCounter(storage Repository, key string) *Counter {
	return &Counter{storage, key}
}

func (a Counter) Update(name string, value string) error {
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf(`counter: %w`, err)
	}
	a.IncrInt64(a.key, name, number)
	return nil
}

func (a Counter) GetStorageValue(name string) (any, bool) {
	return a.Get(a.key, name)
}
