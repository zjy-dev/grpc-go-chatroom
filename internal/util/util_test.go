//go:build unit_test

package util

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWrapGRPCError(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name     string
		err      error
		code     codes.Code
		msg      string
		expected error
	}{
		{
			name:     "nil error",
			err:      nil,
			code:     codes.Internal,
			msg:      "internal error",
			expected: nil,
		},
		{
			name:     "existing grpc error",
			err:      status.Error(codes.NotFound, "not found"),
			code:     codes.Internal,
			msg:      "internal error",
			expected: status.Error(codes.NotFound, "not found"),
		},
		{
			name:     "non-grpc error",
			err:      errors.New("some error"),
			code:     codes.Internal,
			msg:      "internal error",
			expected: status.Error(codes.Internal, "internal error: some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := WrapGRPCError(tt.err, tt.code, tt.msg)
			require.Equal(tt.expected, actual)
		})
	}
}

func TestMustLoadEnvFile(t *testing.T) {
	require := require.New(t)

	// Backup original .env file and environment variable
	originalEnvFile := ".env"
	defer func() {
		// Restore the original .env file
		os.Remove(originalEnvFile)
	}()

	// Test case: .env file exists
	t.Run("env file exists", func(t *testing.T) {

		// Create a temporary .env file
		err := os.WriteFile(originalEnvFile, []byte("TEST_KEY=VALUE"), 0644)
		require.NoError(err)

		require.NotPanics(func() {
			MustLoadEnvFile()
		})
	})

	// Test case: .env file does not exist
	t.Run("env file does not exist", func(t *testing.T) {

		// Remove the .env file
		os.Remove(originalEnvFile)

		require.Panics(func() {
			MustLoadEnvFile()
		})
	})
}

func TestHashPassword(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name     string
		password string
	}{
		{name: "simple password", password: "password123"},
		{name: "empty password", password: ""},
		{name: "long password", password: "thisisaverylongpasswordwithmultiplecharactersandnumbers1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			require.NoError(err)
			require.NotEmpty(hash)
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "correct password",
			password: "password123",
			hash:     func() string { h, _ := HashPassword("password123"); return h }(),
			expected: true,
		},
		{
			name:     "incorrect password",
			password: "password123",
			hash:     func() string { h, _ := HashPassword("password456"); return h }(),
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     func() string { h, _ := HashPassword(""); return h }(),
			expected: true,
		},
		{
			name:     "incorrect hash",
			password: "password123",
			hash:     "$2a$10$invalidhashvalue",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := CheckPasswordHash(tt.password, tt.hash)
			require.Equal(tt.expected, actual)
			if !tt.expected {
				return
			}
			hashedAgain, err := HashPassword(tt.password)
			require.NoError(err)
			require.NotEqual(tt.hash, hashedAgain)
			require.True(CheckPasswordHash(tt.password, hashedAgain))
		})
	}
}
