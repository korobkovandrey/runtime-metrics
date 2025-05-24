package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	err := os.Setenv("ADDRESS", "test_ADDRESS")
	require.NoError(t, err)
	err = os.Setenv("FILE_STORAGE_PATH", "test_FILE_STORAGE_PATH")
	require.NoError(t, err)
	err = os.Setenv("DATABASE_DSN", "test_DATABASE_DSN")
	require.NoError(t, err)
	err = os.Setenv("RESTORE", "true")
	require.NoError(t, err)
	err = os.Setenv("STORE_INTERVAL", "5")
	require.NoError(t, err)
	err = os.Setenv("KEY", "test_KEY")
	require.NoError(t, err)
	err = os.Setenv("PPROF", "true")
	require.NoError(t, err)
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
