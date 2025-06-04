package config

import (
	"testing"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Setenv("ADDRESS", "test.host:1234")
	t.Setenv("POLL_INTERVAL", "3")
	t.Setenv("REPORT_INTERVAL", "11")
	t.Setenv("KEY", "test_KEY")
	t.Setenv("RATE_LIMIT", "15")
	t.Setenv("BATCHING", "true")
	t.Setenv("PPROF_ADDRESS", ":6066")
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
