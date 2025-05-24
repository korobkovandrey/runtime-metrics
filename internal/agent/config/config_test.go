package config

import (
	"os"
	"testing"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	err := os.Setenv("ADDRESS", "test.host:1234")
	require.NoError(t, err)
	err = os.Setenv("POLL_INTERVAL", "3")
	require.NoError(t, err)
	err = os.Setenv("REPORT_INTERVAL", "11")
	require.NoError(t, err)
	err = os.Setenv("KEY", "test_KEY")
	require.NoError(t, err)
	err = os.Setenv("RATE_LIMIT", "15")
	require.NoError(t, err)
	err = os.Setenv("BATCHING", "true")
	require.NoError(t, err)
	err = os.Setenv("PPROF_ADDRESS", ":6066")
	require.NoError(t, err)
	cfg, err := NewConfig()
	require.NoError(t, err)
	assert.Equal(t, "test.host:1234", cfg.Addr)
	assert.Equal(t, 3, cfg.PollInterval)
	assert.Equal(t, 11, cfg.ReportInterval)
	assert.Equal(t, "test_KEY", cfg.Key)
	assert.Equal(t, 15, cfg.RateLimit)
	assert.True(t, cfg.Batching)
	assert.Equal(t, ":6066", cfg.PprofAddr)
	assert.Equal(t, sender.Config{
		UpdateURL:   "http://" + cfg.Addr + "/update/",
		UpdatesURL:  "http://" + cfg.Addr + "/updates/",
		RetryDelays: []time.Duration{time.Second, 3 * time.Second, 5 * time.Second},
		Timeout:     10 * time.Second,
		Key:         []byte(cfg.Key),
		RateLimit:   cfg.RateLimit,
	}, *cfg.Sender)
}
