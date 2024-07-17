package main

import (
	"context"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zjy-dev/grpc-go-chatroom/internal/jwt"
	pb "github.com/zjy-dev/grpc-go-chatroom/internal/proto"

	"google.golang.org/grpc"
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
	chatServer := newChatServer()

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
				_, _ = chatServer.LogIn(context.Background(), &pb.LoginReq{Username: "testuser"})
			},
			wantCode: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setup != nil {
				tt.setup()
			}

			resp, err := chatServer.LogIn(context.Background(), &pb.LoginReq{Username: tt.username})
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
		setupFunc  func(cs *chatServer) context.Context
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "Successful Logout",
			username: "testuser",
			setupFunc: func(cs *chatServer) context.Context {
				cs.clientsMap["testuser"] = nil
				return context.WithValue(context.Background(), "username", "testuser")
			},
			wantErr: false,
		},
		{
			name:     "Invalid Auth Token",
			username: "",
			setupFunc: func(cs *chatServer) context.Context {
				return context.Background()
			},
			wantErr:    true,
			wantErrMsg: "invalid auth token",
		},
		{
			name:     "User Not Found",
			username: "nonexistent",
			setupFunc: func(cs *chatServer) context.Context {
				return context.WithValue(context.Background(), "username", "nonexistent")
			},
			wantErr:    true,
			wantErrMsg: "user: nonexistent not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			cs := newChatServer()
			ctx := tt.setupFunc(cs)

			_, err := cs.LogOut(ctx, &pb.Empty{})

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

type mockChatServerStream struct {
	grpc.ServerStream
	recvMsgs  []*pb.Message
	sendMsgs  []*pb.Message
	recvIndex int
	sendMu    sync.Mutex

	tokenValid bool
}

func (m *mockChatServerStream) Context() context.Context {
	if !m.tokenValid {
		return context.Background()
	}
	return context.WithValue(context.Background(), "username", "testuser")
}

func (m *mockChatServerStream) Send(msg *pb.Message) error {
	m.sendMsgs = append(m.sendMsgs, msg)
	return nil
}

func (m *mockChatServerStream) Recv() (*pb.Message, error) {
	if m.recvIndex >= len(m.recvMsgs) {
		return nil, io.EOF
	}
	msg := m.recvMsgs[m.recvIndex]
	m.recvIndex++
	return msg, nil
}

func TestChat(t *testing.T) {
	tests := []struct {
		name        string
		recvMsgs    []*pb.Message
		setupFunc   func(cs *chatServer)
		wantErr     bool
		wantErrCode codes.Code
		tokenValid  bool
	}{
		{
			name: "SuccessfulChat",
			recvMsgs: []*pb.Message{
				{Text: "Hello"},
				{Text: "How are you?"},
			},
			setupFunc: func(cs *chatServer) {
				cs.clientsMap["testuser"] = nil
			},
			wantErr:    false,
			tokenValid: true,
		},
		{
			name:     "InvalidAuthToken",
			recvMsgs: []*pb.Message{},
			setupFunc: func(cs *chatServer) {
				// No setup needed
			},
			wantErr:     true,
			wantErrCode: codes.Unauthenticated,
			tokenValid:  false,
		},
		{
			name:     "UserNotLoggedIn",
			recvMsgs: []*pb.Message{},
			setupFunc: func(cs *chatServer) {
				// No setup needed
			},
			wantErr:     true,
			wantErrCode: codes.NotFound,
			tokenValid:  true,
		},
		{
			name: "NilMessage",
			recvMsgs: []*pb.Message{
				{Text: "Hello"},
				nil,
			},
			setupFunc: func(cs *chatServer) {
				cs.clientsMap["testuser"] = nil
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
			tokenValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			cs := newChatServer()
			tt.setupFunc(cs)

			stream := &mockChatServerStream{
				recvMsgs:   tt.recvMsgs,
				tokenValid: tt.tokenValid,
			}

			err := cs.Chat(stream)

			if tt.wantErr {
				require.Error(err)
				st, ok := status.FromError(err)
				require.True(ok)
				require.Equal(tt.wantErrCode, st.Code())
			} else {
				require.NoError(err)
				// TODO
				// time.Sleep(time.Second * 2)
				// require.Len(stream.sendMsgs, len(tt.recvMsgs))

				// TODO: check order
				// for i, msg := range tt.recvMsgs {
				// 	if msg != nil {
				// 		require.Equal("testuser: "+msg.Text, stream.sendMsgs[i].Text)
				// 	}
				// }
			}
		})
	}
}
