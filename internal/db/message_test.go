//go:build unit_test

package db

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	pb "github.com/zjy-dev/grpc-go-chatroom/api/chat/v1"
)

func TestInsertMessage(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name         string
		userID       int
		username     string
		message      string
		mockBehavior func(mock sqlmock.Sqlmock)
		expectedID   int64
		expectErr    bool
	}{
		{
			name:     "Successful Insert",
			userID:   1,
			username: "testuser",
			message:  "Hello, World!",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO `messages` \\(user_id, username, message\\) VALUES \\(\\?, \\?, \\?\\);").
					WithArgs(1, "testuser", "Hello, World!").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedID: 1,
			expectErr:  false,
		},
		{
			name:     "Insert Failure",
			userID:   1,
			username: "testuser",
			message:  "Hello, World!",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO `messages` \\(user_id, username, message\\) VALUES \\(\\?, \\?, \\?\\);").
					WithArgs(1, "testuser", "Hello, World!").
					WillReturnError(errors.New("insert failed"))
			},
			expectedID: 0,
			expectErr:  true,
		},
		{
			name:     "Get Last Insert ID Failure",
			userID:   1,
			username: "testuser",
			message:  "Hello, World!",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO `messages` \\(user_id, username, message\\) VALUES \\(\\?, \\?, \\?\\);").
					WithArgs(1, "testuser", "Hello, World!").
					WillReturnResult(sqlmock.NewResult(1, 1)).
					WillReturnError(errors.New("failed to get last inserted message ID"))
			},
			expectedID: 0,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database and defer closing it
			db, mock, err := sqlmock.New()
			require.NoError(err)
			defer db.Close()

			// Set up mock behavior
			tt.mockBehavior(mock)

			// Call the function
			id, err := InsertMessage(db, tt.userID, tt.username, tt.message)

			// Assertions
			if tt.expectErr {
				require.Error(err)
				require.Equal(int64(0), id)
			} else {
				require.NoError(err)
				require.Equal(tt.expectedID, id)
			}

			// Ensure all expectations were met
			require.NoError(mock.ExpectationsWereMet())
		})
	}
}

func TestGetMessages(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name         string
		mockBehavior func(mock sqlmock.Sqlmock)
		expected     []*pb.Message
		expectErr    bool
	}{
		{
			name: "Successful Query",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "username", "message", "created_at"}).
					AddRow(1, 1, "testuser", "Hello, World!", "2024-08-15 00:00:00")
				mock.ExpectQuery("SELECT id, user_id, username, message, created_at FROM `messages`;").
					WillReturnRows(rows)
			},
			expected: []*pb.Message{
				{
					TextContent:   "Hello, World!",
					Username:      "testuser",
					MessageNumber: 1,
				},
			},
			expectErr: false,
		},
		{
			name: "Query Failure",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, user_id, username, message, created_at FROM `messages`;").
					WillReturnError(errors.New("query failed"))
			},
			expected:  nil,
			expectErr: true,
		},
		{
			name: "Row Scan Failure",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "username", "message", "created_at"}).
					AddRow(1, 1, "testuser", "Hello, World!", "2024-08-15 00:00:00")
				mock.ExpectQuery("SELECT id, user_id, username, message FROM `messages`;").
					WillReturnRows(rows)
			},
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database and defer closing it
			db, mock, err := sqlmock.New()
			require.NoError(err)
			defer db.Close()

			// Set up mock behavior
			tt.mockBehavior(mock)

			// Call the function
			messages, err := GetMessages(db)

			// Assertions
			if tt.expectErr {
				require.Error(err)
				require.Nil(messages)
			} else {
				require.NoError(err)
				require.Equal(tt.expected, messages)

				// Ensure all expectations were met
				require.NoError(mock.ExpectationsWereMet())
			}

		})
	}
}
