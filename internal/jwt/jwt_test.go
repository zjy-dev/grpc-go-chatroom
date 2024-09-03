//go:build unit_test

package jwt_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zjy-dev/grpc-go-chatroom/internal/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGenerateJwt(t *testing.T) {
	require := require.New(t)

	// Set up test cases
	tests := []struct {
		name        string
		username    string
		expectedErr error
	}{
		{
			name:        "valid username",
			username:    "testuser",
			expectedErr: nil,
		},
		{
			name:        "empty username",
			username:    "",
			expectedErr: status.Errorf(codes.InvalidArgument, "username is empty"),
		},
	}

	// Execute test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate JWT token
			token, err := jwt.GenerateJwt(tt.username)

			// Verify the result
			if tt.expectedErr != nil {
				require.Error(err)
				require.EqualError(err, tt.expectedErr.Error())
				require.Empty(token)
			} else {
				require.NoError(err)
				require.NotEmpty(token)

				// Verify the token by parsing it
				claims, err := jwt.ParseJwt(token)
				require.NoError(err)
				require.Equal(tt.username, claims.Subject)
			}
		})
	}
}

func TestParseJwt(t *testing.T) {
	require := require.New(t)

	// Set up test cases
	tests := []struct {
		name        string
		tokenString string
		expectedErr error
		username    string
	}{
		{
			name:        "valid token",
			tokenString: func() string { token, _ := jwt.GenerateJwt("testuser"); return token }(),
			expectedErr: nil,
			username:    "testuser",
		},
		{
			name:        "empty token",
			tokenString: "",
			expectedErr: status.Errorf(codes.Unauthenticated, "token is empty"),
		},
		{
			name:        "invalid token",
			tokenString: "invalidtoken",
			expectedErr: status.Errorf(codes.Unauthenticated, "failed to parse token"),
		},
	}

	// Execute test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse JWT token
			claims, err := jwt.ParseJwt(tt.tokenString)

			// Verify the result
			if tt.expectedErr != nil {
				require.Error(err)
				require.Contains(err.Error(), tt.expectedErr.Error())
				require.Nil(claims)
			} else {
				require.NoError(err)
				require.NotNil(claims)
				require.Equal(tt.username, claims.Subject)
			}
		})
	}
}
