package repository

import "errors"

type Store struct {
	data map[Type]Repository
}

func (s Store) Get(strType string) (Repository, error) {
	t := Type(strType)
	if s.data[t] == nil {
		return nil, errors.New(`type is not valid`)
	}
	return s.data[t], nil
}

func NewStore(metrics ...Repository) *Store {
	store := &Store{
		map[Type]Repository{},
	}
	for _, m := range metrics {
		store.data[m.GetType()] = m
	}
	return store
}
