package pbsender

import (
	"context"
	"errors"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/pkg/sign"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func getRealIPInterceptor(realIP string) func(ctx context.Context, method string, req interface{},
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req interface{},
		reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {
		return invoker(metadata.AppendToOutgoingContext(ctx, "XRealIP", realIP), method, req, reply, cc, opts...)
	}
}

func getHashInterceptor(key []byte) func(ctx context.Context, method string, req interface{},
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req interface{},
		reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {
		reqM, ok := req.(proto.Message)
		if !ok {
			return errors.New("request is not a proto message")
		}
		dataBytes, err := proto.Marshal(reqM)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		return invoker(metadata.AppendToOutgoingContext(ctx, "HashSHA256", sign.MakeToString(dataBytes, key)),
			method, req, reply, cc, opts...)
	}
}
