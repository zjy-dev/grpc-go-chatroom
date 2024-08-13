//go:build unit_test

package jwt

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLoadJwtKey(t *testing.T) {
	require := require.New(t)
	testJWTKey := "zjy-dev"

	require.Panics(mustLoadJwtKey)

	// Test when JWT_KEY environment variable is set
	os.Setenv("JWT_KEY", testJWTKey)
	mustLoadJwtKey()
	require.Equal(testJWTKey, jwtKey)

	// Test when JWT_KEY environment variable is not set
	os.Unsetenv("JWT_KEY")
	require.Equal("", os.Getenv("JWT_KEY"))
	jwtKey = ""
}

func mustLoadJwtKeyHelper(t *testing.T) {
	t.Helper()
	testJWTKey := "zjy-dev"
	os.Setenv("JWT_KEY", testJWTKey)
	mustLoadJwtKey()
	require.Equal(t, testJWTKey, jwtKey)
}

// TODO: ADD TEST CASEs

func TestGenerateJwt(t *testing.T) {
	mustLoadJwtKeyHelper(t)
	require := require.New(t)
	t.Cleanup(func() {
		os.Unsetenv("JWT_KEY")
		require.Equal("", os.Getenv("JWT_KEY"))
	})
	tests := []struct {
		name     string
		username string
		wantCode codes.Code
	}{
		{
			name:     "ValidUsername",
			username: "testuser",
			wantCode: codes.OK,
		},
		{
			name:     "EmptyUsername",
			username: "",
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJwt(tt.username)
			if tt.wantCode != codes.OK {
				require.Error(err)
				require.Equal(tt.wantCode, status.Code(err))
			} else {
				require.NoError(err)
				require.NotEmpty(t, token)
			}
		})
	}
}

func TestParseJwt(t *testing.T) {
	mustLoadJwtKeyHelper(t)
	require := require.New(t)
	t.Cleanup(func() {
		os.Unsetenv("JWT_KEY")
		require.Equal("", os.Getenv("JWT_KEY"))
	})
	username := "zjy-dev"
	validToken, err := GenerateJwt(username)
	require.NoError(err)

	tests := []struct {
		name     string
		token    string
		wantCode codes.Code
	}{
		{
			name:     "ValidToken",
			token:    validToken,
			wantCode: codes.OK,
		},
		{
			name:     "EmptyToken",
			token:    "",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "InvalidToken",
			token:    "invalidtoken",
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseJwt(tt.token)
			if tt.wantCode != codes.OK {
				require.Error(err)
				require.Nil(claims)
			} else {
				require.NoError(err)
				require.NotNil(claims)
				require.Equal(username, claims.Subject)
			}
		})
	}
}
