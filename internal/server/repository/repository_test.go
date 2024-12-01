package repository

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server/adapter"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository/memstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

type StoreTest interface {
	Get(t string) (Adapter, error)
}

func TestNewStore(t *testing.T) {
	s := NewStore(memstorage.NewMemStorage())
	assert.IsType(t, s, &Store{})
	assert.Implements(t, (*StoreTest)(nil), s)
}

type testAdapter struct {
	adapter.Repository
}

func (t testAdapter) Update(name string, value string) error {
	return nil
}

func (t testAdapter) GetStorageValue(name string) (any, bool) {
	return nil, false
}

func TestStore_addAdapter(t *testing.T) {
	m := memstorage.NewMemStorage()
	s := NewStore(m)
	const key = `test`
	s.addAdapter(key, testAdapter{m})
	require.Contains(t, s.data, key)
	assert.Implements(t, (*Adapter)(nil), s.data[key])
	assert.Implements(t, (*adapter.Repository)(nil), s.data[key])
}

func TestNewStoreMemStorage(t *testing.T) {
	s := NewStoreMemStorage()
	assert.IsType(t, s, &Store{})
	assert.Implements(t, (*StoreTest)(nil), s)
	require.Contains(t, s.data, gaugeType)
	require.Contains(t, s.data, counterType)
	assert.Implements(t, (*Adapter)(nil), s.data[gaugeType])
	assert.Implements(t, (*adapter.Repository)(nil), s.data[gaugeType])
	assert.Implements(t, (*Adapter)(nil), s.data[counterType])
	assert.Implements(t, (*adapter.Repository)(nil), s.data[counterType])
}

func TestStore_Get(t *testing.T) {
	s := NewStoreMemStorage()
	_, err := s.Get(`test`)
	assert.Error(t, err)
	types := []string{gaugeType, counterType}
	for _, typ := range types {
		a, err := s.Get(typ)
		require.NoError(t, err)
		assert.Implements(t, (*Adapter)(nil), a)
		assert.Implements(t, (*adapter.Repository)(nil), a)
	}
}
