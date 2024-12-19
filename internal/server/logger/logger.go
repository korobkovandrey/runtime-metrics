package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var zapLogger = zap.NewNop()

func Initialize() error {
	var err error
	zapLogger, err = zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("logger.Initialize: %w", err)
	}
	return nil
}

func Sugar() *zap.SugaredLogger {
	return zapLogger.Sugar()
}
