//go:build unit_test

package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zjy-dev/grpc-go-chatroom/internal/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthFunc(t *testing.T) {
	require := require.New(t)

	// Load the JWT key from environment variable
	os.Setenv("JWT_KEY", "mysecretkey")
	t.Cleanup(func() {
		os.Unsetenv("JWT_KEY")
		require.Equal("", os.Getenv("JWT_KEY"))
	})

	tests := []struct {
		name    string
		token   string
		md      map[string][]string
		wantErr bool
		errCode codes.Code
	}{
		{
			name:    "ValidToken",
			md:      map[string][]string{"authorization": {"bearer " + createTestToken("testuser")}},
			wantErr: false,
			errCode: codes.OK,
		},
		{
			name:    "InvalidTokenPrefix",
			md:      map[string][]string{"authorization": {"invalid " + createTestToken("testuser")}},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
		{
			name:    "EmptyToken",
			md:      map[string][]string{"authorization": {"bearer "}},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
		{
			name:    "InvalidToken",
			md:      map[string][]string{"authorization": {"bearer invalid-token"}},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := metadata.NewIncomingContext(context.Background(), tt.md)
			_, err := authFunc(ctx)

			if tt.wantErr {
				require.Error(err)
				st, _ := status.FromError(err)
				require.Equal(tt.errCode, st.Code())
			} else {
				require.NoError(err)
			}
		})
	}
}

func createTestToken(username string) string {
	token, _ := jwt.GenerateJwt(username)
	return token
}
