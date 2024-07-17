package middlewares

import (
	"context"
	"strings"

	"google.golang.org/grpc"
)

type AuthFunc func(ctx context.Context) (context.Context, error)

// UnaryServerAuthInterceptorWithBypassMethods creates a new unary server interceptor.
// The interceptor will call the authFunc only if the request's method name
// doesn't match any of the bypassMethods. The bypassMethods are matched against
// the full method name (e.g. /service.Service/Method).
//
// Returns a grpc.UnaryServerInterceptor.
func UnaryServerAuthInterceptorWithBypassMethods(authFunc AuthFunc, bypassMethods ...string) grpc.UnaryServerInterceptor {
	// This function is the actual interceptor.
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Check if the method is in the bypass list.
		for _, method := range bypassMethods {
			if strings.HasSuffix(info.FullMethod, method) {
				// If it is, call the handler directly without calling authFunc.
				return handler(ctx, req)
			}
		}

		// If the method isn't in the bypass list, call authFunc and then call the handler.
		ctx, err := authFunc(ctx)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}
