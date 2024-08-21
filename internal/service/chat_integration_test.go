//go:build integration_test

package service

import (
	"context"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	pb "github.com/zjy-dev/grpc-go-chatroom/api/chat/v1"

	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	os.Setenv("DBUSER", "root")
	os.Setenv("DBPASS", "123456")
	os.Setenv("JWT_KEY", "zjy-dev")

	code := m.Run()

	os.Unsetenv("DBUSER")
	os.Unsetenv("DBPASS")
	os.Unsetenv("JWT_KEY")

	os.Exit(code)
}

type mockChatServerStream struct {
	grpc.ServerStream
	requests          []*pb.ChatRequest
	responses         []*pb.ChatResponse
	reqIndex          int
	expectResponseLen int
	totolUsersNumber  int
	mux               sync.Mutex
	username          string
	cs                *chatServiceServer
}

func (m *mockChatServerStream) Context() context.Context {
	if len(m.username) == 0 {
		return context.Background()
	}
	return context.WithValue(context.Background(), JWTContextKey, m.username)
}

func (m *mockChatServerStream) Send(resp *pb.ChatResponse) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.responses = append(m.responses, resp)
	return nil
}

func (m *mockChatServerStream) Recv() (*pb.ChatRequest, error) {
	if m.reqIndex >= len(m.requests) {
		m.mux.Lock()
		for len(m.responses) < m.expectResponseLen {
			m.mux.Unlock()
			time.Sleep(time.Millisecond * 100)
			m.mux.Lock()
		}
		m.mux.Unlock()
		return nil, io.EOF
	}
	if m.reqIndex == 0 {
		m.cs.mu.Lock()
		for len(m.cs.clientsMap) < m.totolUsersNumber {
			m.cs.mu.Unlock()
			time.Sleep(time.Millisecond * 100)
			m.cs.mu.Lock()
		}
		m.cs.mu.Unlock()
	}
	req := m.requests[m.reqIndex]
	m.reqIndex++
	return req, nil
}

func TestChatIntegration(t *testing.T) {
	require := require.New(t)

	t.Run("TwoUsers", func(t *testing.T) {
		cs := NewChatServiceServer()
		cs.clientsMap["user1"] = client{}
		cs.clientsMap["user2"] = client{}

		stream1 := &mockChatServerStream{
			cs:                cs,
			totolUsersNumber:  2,
			expectResponseLen: 2,
			username:          "user1",
			requests: []*pb.ChatRequest{
				{Message: &pb.Message{Type: pb.MessageType_MESSAGE_TYPE_NORMAL, TextContent: "user1-1"}},
				{Message: &pb.Message{Type: pb.MessageType_MESSAGE_TYPE_NORMAL, TextContent: "user1-2"}},
				{Message: &pb.Message{Type: pb.MessageType_MESSAGE_TYPE_NORMAL, TextContent: "user1-3"}},
			},
		}
		stream2 := &mockChatServerStream{
			cs:                cs,
			totolUsersNumber:  2,
			expectResponseLen: 3,
			username:          "user2",
			requests: []*pb.ChatRequest{
				{Message: &pb.Message{Type: pb.MessageType_MESSAGE_TYPE_NORMAL, TextContent: "user2-1"}},
				{Message: &pb.Message{Type: pb.MessageType_MESSAGE_TYPE_NORMAL, TextContent: "user2-2"}},
			},
		}
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			require.Nil(cs.Chat(stream1))
		}()
		go func() {
			defer wg.Done()
			require.Nil(cs.Chat(stream2))
		}()
		wg.Wait()
		require.Equal(stream1.expectResponseLen, len(stream1.responses))
		require.Equal(stream2.expectResponseLen, len(stream2.responses))
		require.Equal(len(cs.clientsMap), 0)
		require.Equal("user2", stream1.responses[0].GetMessage().GetUsername())
		require.Equal("user1", stream2.responses[0].GetMessage().GetUsername())
		require.Equal("user2-1", stream1.responses[0].GetMessage().GetTextContent())
		require.Equal("user1-1", stream2.responses[0].GetMessage().GetTextContent())
	})
}
