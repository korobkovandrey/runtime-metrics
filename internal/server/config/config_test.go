package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Setenv("ADDRESS", "test_ADDRESS")
	t.Setenv("FILE_STORAGE_PATH", "test_FILE_STORAGE_PATH")
	t.Setenv("DATABASE_DSN", "test_DATABASE_DSN")
	t.Setenv("RESTORE", "true")
	t.Setenv("STORE_INTERVAL", "5")
	t.Setenv("KEY", "test_KEY")
	t.Setenv("PPROF", "true")
	cfg, err := NewConfig()
	require.NoError(t, err)
	assert.Equal(t, "test_ADDRESS", cfg.Addr)
	assert.Equal(t, "test_FILE_STORAGE_PATH", cfg.FileStoragePath)
	assert.Equal(t, "test_DATABASE_DSN", cfg.DatabaseDSN)
	assert.True(t, cfg.Restore)
	assert.Equal(t, int64(5), cfg.StoreInterval)
	assert.Equal(t, "test_KEY", cfg.Key)
	assert.True(t, cfg.Pprof)
	assert.Equal(t, 5*time.Second, cfg.ShutdownTimeout)
	assert.Equal(t, 5*time.Second, cfg.DatabasePingTimeout)
	assert.Equal(t, []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}, cfg.RetryDelays)
}
