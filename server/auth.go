package main

import (
	"context"

	authmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/zjy-dev/grpc-go-chatroom/internal/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// authFunc is a function that authenticates incoming requests.
func authFunc(ctx context.Context) (context.Context, error) {
	// Get the token from the metadata.
	token, err := authmiddleware.AuthFromMD(ctx, "bearer")

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid auth token prefix: %v", err)
	}

	// Parse the token.
	claims, err := jwt.ParseJwt(token)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parse token failed: %v", err)
	}
	// Get the subject from the claims.
	subject, err := claims.GetSubject()

	// This err is forever nil due to the design
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get subject failed: %v", err)
	}

	if subject == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is empty")
	}

	// TODO:
	return context.WithValue(ctx, "username", subject), nil
}
