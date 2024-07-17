package tokensource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{name: "empty token", token: ""},
		{name: "non-empty token", token: "test_token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			auth := New(tt.token)
			require.Equal(tt.token, auth.token)
		})
	}
}

func TestGetRequestMetadata(t *testing.T) {
	tests := []struct {
		name   string
		auth   Auth
		ctx    context.Context
		uri    []string
		result map[string]string
		err    error
	}{
		{
			name: "empty token",
			auth: Auth{token: ""},
			ctx:  context.Background(),
			uri:  nil,
			result: map[string]string{
				"authorization": "bearer ",
			},
			err: nil,
		},
		{
			name: "non-empty token",
			auth: Auth{token: "test_token"},
			ctx:  context.Background(),
			uri:  nil,
			result: map[string]string{
				"authorization": "bearer test_token",
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			result, err := tt.auth.GetRequestMetadata(tt.ctx, tt.uri...)
			require.Equal(tt.result, result)
			require.Equal(tt.err, err)
		})
	}
}

func TestRequireTransportSecurity(t *testing.T) {
	tests := []struct {
		name string
		auth Auth
		want bool
	}{
		{
			name: "any token",
			auth: Auth{token: "any_token"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			result := tt.auth.RequireTransportSecurity()
			require.Equal(tt.want, result)
		})
	}
}
