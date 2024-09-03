//go:build unit_test

package logic

import (
	"context"
	"database/sql"
	"io"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	pb "github.com/zjy-dev/grpc-go-chatroom/api/chat/v1"
	"github.com/zjy-dev/grpc-go-chatroom/internal/jwt"
	"github.com/zjy-dev/grpc-go-chatroom/internal/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock database setup
func mockDB() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return db, mock
}

func TestMain(m *testing.M) {
	os.Setenv("JWT_KEY", "zjy-dev")

	code := m.Run()

	os.Unsetenv("JWT_KEY")

	os.Exit(code)
}
func TestLogInOrRegister(t *testing.T) {
	require := require.New(t)

	type args struct {
		req *pb.LogInOrRegisterRequest
	}
	tests := []struct {
		name            string
		args            args
		alreadyLoggedIn bool
		mockSetup       func(mock sqlmock.Sqlmock)
		expectedError   error
	}{
		{
			name: "successful registration",
			args: args{
				req: &pb.LogInOrRegisterRequest{
					Username: "newuser",
					Password: "password123",
				},
			},
			alreadyLoggedIn: false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM `users` WHERE username = ?").
					WithArgs("newuser").
					WillReturnRows(sqlmock.NewRows([]string{"id"})) // No rows returned

				mock.ExpectExec("INSERT INTO `users` \\(`username`, `password_hash`\\) VALUES \\(\\?, \\?\\);").
					WithArgs("newuser", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: nil,
		},
		{
			name: "already logged in",
			args: args{
				req: &pb.LogInOrRegisterRequest{
					Username: "existinguser",
					Password: "password123",
				},
			},
			alreadyLoggedIn: true,
			mockSetup: func(mock sqlmock.Sqlmock) {
			},
			expectedError: status.Errorf(codes.AlreadyExists, "user: %s has already logged in", "existinguser"),
		},
		{
			name: "user already registered and password matches",
			args: args{
				req: &pb.LogInOrRegisterRequest{
					Username: "existinguser",
					Password: "password123",
				},
			},
			alreadyLoggedIn: false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM `users` WHERE username = ?").
					WithArgs("existinguser").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				hashedPassword, err := util.HashPassword("password123")
				require.NoError(err)
				mock.ExpectQuery("SELECT id, username, password_hash FROM `users` WHERE username = ?").
					WithArgs("existinguser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash"}).
						AddRow(1, "existinguser", hashedPassword))
			},
			expectedError: nil,
		},
		{
			name: "user already registered but password does not match",
			args: args{
				req: &pb.LogInOrRegisterRequest{
					Username: "existinguser",
					Password: "wrongpassword",
				},
			},
			alreadyLoggedIn: false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM `users` WHERE username = ?").
					WithArgs("existinguser").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				hashedPassword, _ := util.HashPassword("password123") // Correct password
				mock.ExpectQuery("SELECT id, username, password_hash FROM `users` WHERE username = ?").
					WithArgs("existinguser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash"}).
						AddRow(1, "existinguser", hashedPassword))
			},
			expectedError: status.Errorf(codes.Unauthenticated, "incorrect password"),
		},
		{
			name: "invalid username or password length",
			args: args{
				req: &pb.LogInOrRegisterRequest{
					Username: "a",
					Password: "p",
				},
			},
			mockSetup:     func(mock sqlmock.Sqlmock) {},
			expectedError: status.Errorf(codes.InvalidArgument, "invalid username or password length"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := mockDB()
			defer db.Close()
			tt.mockSetup(mock)

			// Use the mock db connection
			dbConn = db
			defer func() { dbConn = nil }() // reset dbConn after test

			cs := NewChatServiceServer()

			if tt.alreadyLoggedIn {
				cs.clientsMap = map[string]client{
					"existinguser": {},
				}
			}

			resp, err := cs.LogInOrRegister(context.Background(), tt.args.req)
			if tt.expectedError != nil {
				require.Equal(tt.expectedError, err)
			} else {
				require.NoError(err)
				claims, err := jwt.ParseJwt(resp.Token)
				require.NoError(err)
				require.Equal(claims.Subject, tt.args.req.Username)
			}

			require.NoError(mock.ExpectationsWereMet())
		})
	}
}

// TestLogOut tests the LogOut method.
func TestLogOut(t *testing.T) {
	require := require.New(t)

	type args struct {
		ctx context.Context
		req *pb.LogOutRequest
	}
	tests := []struct {
		name          string
		args          args
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedError error
	}{
		{
			name: "successful logout",
			args: args{
				ctx: context.WithValue(context.Background(), JWTContextKey, "existinguser"),
				req: &pb.LogOutRequest{},
			},
			mockSetup:     func(mock sqlmock.Sqlmock) {},
			expectedError: nil,
		},
		{
			name: "user not found",
			args: args{
				ctx: context.WithValue(context.Background(), JWTContextKey, "unknownuser"),
				req: &pb.LogOutRequest{},
			},
			mockSetup:     func(mock sqlmock.Sqlmock) {},
			expectedError: status.Errorf(codes.NotFound, "user: unknownuser not found"),
		},
		{
			name: "invalid auth token",
			args: args{
				ctx: context.Background(),
				req: &pb.LogOutRequest{},
			},
			mockSetup:     func(mock sqlmock.Sqlmock) {},
			expectedError: status.Errorf(codes.Unauthenticated, "invalid auth token"),
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewChatServiceServer()
			cs.clientsMap = map[string]client{
				"existinguser": {messageChan: make(chan *pb.Message)},
			}

			resp, err := cs.LogOut(tt.args.ctx, tt.args.req)
			if tt.expectedError != nil {
				require.Equal(tt.expectedError, err)
			} else {
				require.NoError(err)
				require.NotNil(resp)
			}
		})
	}
}

func TestChat(t *testing.T) {
	require := require.New(t)

	// Create a new sqlmock database connection and mock object
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(err)
	defer db.Close()

	// Mock db connection
	dbConn = db

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

		// Mock InsertMessage calls
		for range 5 {
			mock.ExpectExec("INSERT INTO `messages` (user_id, username, message) VALUES (?, ?, ?);").
				WithArgs(42, sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
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

		// Ensure all expectations were met
		require.NoError(mock.ExpectationsWereMet())
	})
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
