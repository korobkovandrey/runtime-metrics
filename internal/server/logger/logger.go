package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func NewZapLogger() (zapLogger *zap.Logger, err error) {
	zapLogger, err = zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("NewLogger: %w", err)
	}
	return
}
