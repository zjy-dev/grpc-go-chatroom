//go:build unit_test

package service

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	pb "github.com/zjy-dev/grpc-go-chatroom/api/chat/v1"
	"github.com/zjy-dev/grpc-go-chatroom/internal/jwt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mustSetJWTKey(t *testing.T) {
	t.Helper()
	testJWTKey := "zjy-dev"
	os.Setenv("JWT_KEY", testJWTKey)
	require.Equal(t, testJWTKey, os.Getenv("JWT_KEY"))
}

func TestLogIn(t *testing.T) {
	mustSetJWTKey(t)
	require := require.New(t)
	t.Cleanup(func() {
		os.Unsetenv("JWT_KEY")
		require.Equal("", os.Getenv("JWT_KEY"))
	})
	chatServer := NewChatServiceServer()

	tests := []struct {
		name     string
		username string
		setup    func()
		wantCode codes.Code
	}{
		{
			name:     "EmptyUsername",
			username: "",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "SuccessfulLogin",
			username: "testuser",
			wantCode: codes.OK,
		},
		{
			name:     "DuplicateLogin",
			username: "testuser",
			setup: func() {
				_, _ = chatServer.LogIn(context.Background(), &pb.LogInRequest{Username: "testuser"})
			},
			wantCode: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setup != nil {
				tt.setup()
			}

			resp, err := chatServer.LogIn(context.Background(), &pb.LogInRequest{Username: tt.username})
			if tt.wantCode == codes.OK {
				require.NoError(err)
				require.NotNil(resp)
				token, err := jwt.ParseJwt(resp.Token)
				require.NoError(err)
				sub, err := token.GetSubject()
				require.NoError(err)
				require.Equal(tt.username, sub)
			} else {
				require.Nil(resp)
				require.Error(err)
				require.Equal(tt.wantCode, status.Code(err))
			}
		})
	}
}

func TestLogOut(t *testing.T) {
	tests := []struct {
		name       string
		username   string
		setupFunc  func(cs *chatServiceServer) context.Context
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "Successful Logout With Already Chatted",
			username: "testuser",
			setupFunc: func(cs *chatServiceServer) context.Context {
				cs.clientsMap["testuser"] = client{messageChan: make(chan *pb.Message, 1<<3)}
				return context.WithValue(context.Background(), JWTContextKey, "testuser")
			},
			wantErr: false,
		},
		{
			name:     "Successful Logout Without Chatted",
			username: "testuser",
			setupFunc: func(cs *chatServiceServer) context.Context {
				cs.clientsMap["testuser"] = client{}
				return context.WithValue(context.Background(), JWTContextKey, "testuser")
			},
			wantErr: false,
		},
		{
			name:     "Invalid Auth Token",
			username: "",
			setupFunc: func(cs *chatServiceServer) context.Context {
				return context.Background()
			},
			wantErr:    true,
			wantErrMsg: "invalid auth token",
		},
		{
			name:     "User Not Found",
			username: "nonexistent",
			setupFunc: func(cs *chatServiceServer) context.Context {
				return context.WithValue(context.Background(), JWTContextKey, "nonexistent")
			},
			wantErr:    true,
			wantErrMsg: "user: nonexistent not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			cs := NewChatServiceServer()
			ctx := tt.setupFunc(cs)

			_, err := cs.LogOut(ctx, &pb.LogOutRequest{})

			if tt.wantErr {
				require.Error(err)
				st, _ := status.FromError(err)
				require.Equal(tt.wantErrMsg, st.Message())
			} else {
				require.NoError(err)
				_, ok := cs.clientsMap["testuser"]
				require.False(ok)
			}
		})
	}
}
