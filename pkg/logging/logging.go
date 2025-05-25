// Package logging provides a logging package for the application.
package logging

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type key string

const zapFieldsKey key = "zapFields"

type ZapFields map[string]zap.Field

// Append appends the provided fields to the existing ZapFields.
func (zf ZapFields) Append(fields ...zap.Field) ZapFields {
	zfCopy := make(ZapFields)
	for k, v := range zf {
		zfCopy[k] = v
	}

	for _, f := range fields {
		zfCopy[f.Key] = f
	}

	return zfCopy
}

// ZapLogger is a wrapper around zap.Logger.
type ZapLogger struct {
	logger *zap.Logger
	level  zap.AtomicLevel
}

// NewZapLogger returns a new ZapLogger configured with the provided options.
func NewZapLogger(level zapcore.Level) (*ZapLogger, error) {
	atomic := zap.NewAtomicLevelAt(level)
	settings := defaultSettings(atomic)

	l, err := settings.config.Build(settings.opts...)
	if err != nil {
		return nil, fmt.Errorf("NewZapLogger: %w", err)
	}

	return &ZapLogger{
		logger: l,
		level:  atomic,
	}, nil
}

// WithContextFields returns a new context with the provided fields.
func (z *ZapLogger) WithContextFields(ctx context.Context, fields ...zap.Field) context.Context {
	ctxFields, _ := ctx.Value(zapFieldsKey).(ZapFields)
	if ctxFields == nil {
		ctxFields = make(ZapFields)
	}

	merged := ctxFields.Append(fields...)
	return context.WithValue(ctx, zapFieldsKey, merged)
}

// maskField masks the value of a field if its key is "password" or "email".
func (z *ZapLogger) maskField(f zap.Field) zap.Field {
	if f.Key == "password" {
		return zap.String(f.Key, "******")
	}

	if f.Key == "email" {
		email := f.String
		parts := strings.Split(email, "@")
		if len(parts) == 2 {
			return zap.String(f.Key, "***@"+parts[1])
		}
	}
	return f
}

// Sync flushes any buffered log entries.
func (z *ZapLogger) Sync() {
	_ = z.logger.Sync()
}

// withCtxFields appends the provided fields to the existing ZapFields in the context.
func (z *ZapLogger) withCtxFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	fs := make(ZapFields)

	ctxFields, ok := ctx.Value(zapFieldsKey).(ZapFields)
	if ok {
		fs = ctxFields
	}

	fs = fs.Append(fields...)

	maskedFields := make([]zap.Field, len(fs))
	i := 0
	for _, f := range fs {
		maskedFields[i] = z.maskField(f)
		i++
	}

	return maskedFields
}

// InfoCtx logs an info message with the provided fields.
func (z *ZapLogger) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Info(msg, z.withCtxFields(ctx, fields...)...)
}

// DebugCtx logs a debug message with the provided fields.
func (z *ZapLogger) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Debug(msg, z.withCtxFields(ctx, fields...)...)
}

// WarnCtx logs a warning message with the provided fields.
func (z *ZapLogger) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Warn(msg, z.withCtxFields(ctx, fields...)...)
}

// ErrorCtx logs an error message with the provided fields.
func (z *ZapLogger) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Error(msg, z.withCtxFields(ctx, fields...)...)
}

// FatalCtx logs a fatal message with the provided fields.
func (z *ZapLogger) FatalCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Fatal(msg, z.withCtxFields(ctx, fields...)...)
}

// PanicCtx logs a panic message with the provided fields.
func (z *ZapLogger) PanicCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Panic(msg, z.withCtxFields(ctx, fields...)...)
}

// SetLevel sets the log level.
func (z *ZapLogger) SetLevel(level zapcore.Level) {
	z.level.SetLevel(level)
}

// Std returns a log.Logger that writes to the zap.Logger.
func (z *ZapLogger) Std() *log.Logger {
	return zap.NewStdLog(z.logger)
}

// Logger returns the underlying zap.Logger.
func (z *ZapLogger) Logger() *zap.Logger {
	return z.logger
}
