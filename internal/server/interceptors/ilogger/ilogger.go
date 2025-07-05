package ilogger

import (
	"context"

	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Interceptor(l *logging.ZapLogger) func(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		f := []zap.Field{
			zap.String("method", info.FullMethod),
		}
		if err != nil {
			f = append(f, zap.Error(err))
		}
		l.InfoCtx(ctx, "GRPC request", f...)
		return resp, err
	}
}
