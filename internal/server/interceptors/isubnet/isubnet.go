package isubnet

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Interceptor(subnet *net.IPNet) func(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if subnet == nil {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "missing XRealIP")
		}
		ipStrs := md.Get("XRealIP")
		if len(ipStrs) == 0 || ipStrs[0] == "" {
			return nil, status.Error(codes.InvalidArgument, "missing XRealIP")
		}
		ip := net.ParseIP(ipStrs[0])
		if ip == nil {
			return nil, status.Error(codes.InvalidArgument, "invalid XRealIP")
		}
		if !subnet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "XRealIP is not in the subnet")
		}
		return handler(ctx, req)
	}
}
