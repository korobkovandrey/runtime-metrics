package isign

import (
	"context"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/pkg/sign"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func Interceptor(key []byte) func(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if len(key) == 0 {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "missing HashSHA256")
		}
		tmp := md.Get("HashSHA256")
		if len(tmp) == 0 || tmp[0] == "" {
			return nil, status.Error(codes.InvalidArgument, "missing HashSHA256")
		}
		bh, err := sign.DecodeString(tmp[0])
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to decode HashSHA256: %w", err).Error())
		}
		reqM, ok := req.(proto.Message)
		if !ok {
			return nil, status.Error(codes.Internal, "request is not a proto message")
		}
		dataBytes, err := proto.Marshal(reqM)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("failed to marshal request: %w", err).Error())
		}
		if !sign.Validate(dataBytes, key, bh) {
			return nil, status.Error(codes.PermissionDenied, "invalid signature")
		}
		return handler(ctx, req)
	}
}
