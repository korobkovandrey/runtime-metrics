package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// captureOutput captures the output written to os.Stderr during the execution of f
func captureOutput(t *testing.T, f func()) string {
	originalStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	f()

	_ = w.Close()
	<-done
	os.Stderr = originalStderr
	return buf.String()
}

func TestZapLogger(t *testing.T) {
	t.Run("Log Levels", func(t *testing.T) {
		output := captureOutput(t, func() {
			logger, err := NewZapLogger(zapcore.WarnLevel)
			assert.NoError(t, err)
			ctx := t.Context()
			logger.DebugCtx(ctx, "debug message")
			logger.InfoCtx(ctx, "info message")
			logger.WarnCtx(ctx, "warn message")
			logger.ErrorCtx(ctx, "error message")
			logger.Sync()
		})

		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, lines, 2, "Only Warn and Error logs should be output")
		for _, line := range lines {
			var entry map[string]interface{}
			err := json.Unmarshal([]byte(line), &entry)
			assert.NoError(t, err)
			level := entry["level"].(string)
			assert.True(t, level == "WARN" || level == "ERROR", "Log level should be WARN or ERROR")
		}
	})

	t.Run("Context Fields", func(t *testing.T) {
		output := captureOutput(t, func() {
			logger, err := NewZapLogger(zapcore.InfoLevel)
			assert.NoError(t, err)
			ctx := logger.WithContextFields(t.Context(), zap.String("ctx_key", "ctx_value"))
			logger.InfoCtx(ctx, "test message", zap.String("msg_key", "msg_value"))
			logger.Sync()
		})

		var entry map[string]interface{}
		err := json.Unmarshal([]byte(output), &entry)
		assert.NoError(t, err)
		assert.Equal(t, "INFO", entry["level"])
		assert.Equal(t, "test message", entry["message"])
		assert.Equal(t, "ctx_value", entry["ctx_key"])
		assert.Equal(t, "msg_value", entry["msg_key"])
	})

	t.Run("Sensitive Field Masking", func(t *testing.T) {
		output := captureOutput(t, func() {
			logger, err := NewZapLogger(zapcore.InfoLevel)
			assert.NoError(t, err)
			ctx := t.Context()
			logger.InfoCtx(ctx, "login attempt",
				zap.String("password", "secret"),
				zap.String("email", "user@example.com"),
				zap.String("username", "john"))
			logger.Sync()
		})

		var entry map[string]interface{}
		err := json.Unmarshal([]byte(output), &entry)
		assert.NoError(t, err)
		assert.Equal(t, "******", entry["password"])
		assert.Equal(t, "***@example.com", entry["email"])
		assert.Equal(t, "john", entry["username"])
	})

	t.Run("Dynamic Log Level Change", func(t *testing.T) {
		output := captureOutput(t, func() {
			logger, err := NewZapLogger(zapcore.InfoLevel)
			assert.NoError(t, err)
			ctx := t.Context()
			logger.SetLevel(zapcore.DebugLevel)
			logger.DebugCtx(ctx, "debug message")
			logger.SetLevel(zapcore.ErrorLevel)
			logger.InfoCtx(ctx, "info message")
			logger.ErrorCtx(ctx, "error message")
			logger.Sync()
		})

		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, lines, 2, "Should have Debug and Error logs")
		levels := []string{}
		for _, line := range lines {
			var entry map[string]interface{}
			err := json.Unmarshal([]byte(line), &entry)
			assert.NoError(t, err)
			levels = append(levels, entry["level"].(string))
		}
		assert.Contains(t, levels, "DEBUG")
		assert.Contains(t, levels, "ERROR")
		assert.NotContains(t, levels, "INFO")
	})

	t.Run("Standard Logger", func(t *testing.T) {
		output := captureOutput(t, func() {
			logger, err := NewZapLogger(zapcore.InfoLevel)
			assert.NoError(t, err)
			stdLogger := logger.Std()
			stdLogger.Println("test message")
			logger.Sync()
		})

		var entry map[string]interface{}
		err := json.Unmarshal([]byte(output), &entry)
		assert.NoError(t, err)
		assert.Equal(t, "INFO", entry["level"])
		assert.Contains(t, entry["message"], "test message")
	})

	t.Run("PanicCtx", func(t *testing.T) {
		output := captureOutput(t, func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("PanicCtx did not panic")
				}
			}()
			logger, err := NewZapLogger(zapcore.InfoLevel)
			assert.NoError(t, err)
			ctx := t.Context()
			logger.PanicCtx(ctx, "panic message")
		})

		var entry map[string]interface{}
		err := json.Unmarshal([]byte(output), &entry)
		assert.NoError(t, err)
		assert.Equal(t, "PANIC", entry["level"])
		assert.Equal(t, "panic message", entry["message"])
	})

	t.Run("FatalCtx", func(t *testing.T) {
		//nolint:gosec // Run in a subprocess to capture os.Exit
		cmd := exec.Command(os.Args[0], "-test.run=TestFatalCtxSubprocess")
		cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
		var out bytes.Buffer
		cmd.Stderr = &out
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok {
			assert.Equal(t, 1, e.ExitCode(), "Exit code should be 1")
		} else {
			t.Fatalf("Expected ExitError, got %v", err)
		}

		var entry map[string]interface{}
		err = json.Unmarshal(out.Bytes(), &entry)
		assert.NoError(t, err)
		assert.Equal(t, "FATAL", entry["level"])
		assert.Equal(t, "fatal message", entry["message"])
	})
}

// TestFatalCtxSubprocess is a helper test run in a subprocess for FatalCtx
func TestFatalCtxSubprocess(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") != "1" {
		return
	}
	logger, err := NewZapLogger(zapcore.InfoLevel)
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	logger.FatalCtx(ctx, "fatal message")
}

func TestLogger(t *testing.T) {
	logger, err := NewZapLogger(zapcore.WarnLevel)
	require.NoError(t, err)
	assert.Same(t, logger.logger, logger.Logger())
}
