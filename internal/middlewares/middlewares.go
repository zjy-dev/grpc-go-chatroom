package middlewares

import (
	"context"
	"strings"

	"google.golang.org/grpc"
)

type AuthFunc func(ctx context.Context) (context.Context, error)

func UnaryServerInterceptorWithBypassMethods(authFunc AuthFunc, bypassMethods ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		for _, method := range bypassMethods {
			if strings.HasSuffix(info.FullMethod, method) {
				return handler(ctx, req)
			}
		}

		authFunc(ctx)

		return handler(ctx, req)
	}
}
