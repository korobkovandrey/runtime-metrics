package factory

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryFactory(t *testing.T) {
	ctx := t.Context()

	tests := []struct {
		name          string
		config        *config.Config
		setup         func(t *testing.T) func()
		expectedType  interface{}
		expectedError string
	}{
		{
			name: "PostgreSQL storage with invalid DSN",
			config: &config.Config{
				DatabaseDSN:         "invalid-dsn",
				DatabasePingTimeout: 1 * time.Second,
				RetryDelays:         []time.Duration{100 * time.Millisecond},
			},
			expectedType:  nil,
			expectedError: "failed to create pgxstorage",
		},
		{
			name: "File storage with valid path",
			config: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "storage.json"),
				Restore:         false,
			},
			expectedType:  &repository.FileStorage{},
			expectedError: "",
		},
		{
			name: "File storage with restore and non-existent file",
			config: &config.Config{
				FileStoragePath: filepath.Join(t.TempDir(), "nonexistent.json"),
				Restore:         true,
			},
			expectedType:  &repository.FileStorage{},
			expectedError: "",
		},
		{
			name:          "Memory storage",
			config:        &config.Config{},
			expectedType:  &repository.MemStorage{},
			expectedError: "",
		},
	}
	type closer interface {
		Close() error
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := RepositoryFactory(ctx, tt.config, nil)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tt.expectedType, repo)
				if cl, ok := repo.(closer); ok {
					assert.NoError(t, cl.Close())
				}
			}
		})
	}
}
