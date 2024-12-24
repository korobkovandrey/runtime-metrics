package repository

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server/adapter"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestNewStoreMemStorage(t *testing.T) {
	s, err := NewStoreMemStorage(&config.Config{})
	require.NoError(t, err)
	assert.IsType(t, s, &Store{})
	require.Contains(t, s.data, gaugeType)
	require.Contains(t, s.data, counterType)
	assert.Implements(t, (*Adapter)(nil), s.data[gaugeType])
	assert.Implements(t, (*adapter.Repository)(nil), s.data[gaugeType])
	assert.Implements(t, (*Adapter)(nil), s.data[counterType])
	assert.Implements(t, (*adapter.Repository)(nil), s.data[counterType])
}

func TestStore_Get(t *testing.T) {
	s, err := NewStoreMemStorage(&config.Config{})
	require.NoError(t, err)
	_, err = s.Get("test")
	assert.Error(t, err)
	types := []string{gaugeType, counterType}
	for _, typ := range types {
		a, err := s.Get(typ)
		require.NoError(t, err)
		assert.Implements(t, (*Adapter)(nil), a)
		assert.Implements(t, (*adapter.Repository)(nil), a)
	}
}
