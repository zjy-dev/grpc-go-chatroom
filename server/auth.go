package main

import (
	"context"

	authmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/zjy-dev/grpc-go-chatroom/internal/jwt"
	"github.com/zjy-dev/grpc-go-chatroom/internal/util"
	"github.com/zjy-dev/grpc-go-chatroom/logic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// authFunc is a function that authenticates incoming requests.
func authFunc(ctx context.Context) (context.Context, error) {
	// Get the token from the metadata.
	token, err := authmiddleware.AuthFromMD(ctx, "bearer")

	if err != nil {
		return nil, util.WrapGRPCError(err, codes.Unauthenticated, "invalid auth token prefix")
	}

	// Parse the token.
	claims, err := jwt.ParseJwt(token)
	if err != nil {
		return nil, util.WrapGRPCError(err, codes.Unauthenticated, "parse token failed")
	}
	// Get the subject from the claims. Subject means username in this context.
	subject, err := claims.GetSubject()

	// This err is forever nil due to the design
	if err != nil {
		return nil, util.WrapGRPCError(err, codes.Unauthenticated, "get subject failed")
	}

	if subject == "" {
		return nil, status.Errorf(codes.Unauthenticated, "username in jwt is empty")
	}

	return context.WithValue(ctx, logic.JWTContextKey, subject), nil
}
