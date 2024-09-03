//go:build unit_test

package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestInsertUser(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name        string
		username    string
		password    string
		mockSetup   func(sqlmock.Sqlmock)
		expectedID  int64
		expectedErr error
	}{
		{
			name:     "Successful Insert",
			username: "newuser",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO `users` \\(`username`, `password_hash`\\) VALUES \\(\\?, \\?\\);").
					WithArgs("newuser", "hashedPassword").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name:     "Insert Error",
			username: "newuser",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO `users` \\(`username`, `password_hash`\\) VALUES \\(\\?, \\?\\);").
					WithArgs("newuser", "hashedPassword").
					WillReturnError(errors.New("insert failed"))
			},
			expectedID:  0,
			expectedErr: errors.New("failed to insert user to database: insert failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(err)
			defer db.Close()

			tt.mockSetup(mock)

			id, err := InsertUser(db, tt.username, tt.password)
			require.Equal(tt.expectedID, id)
			require.Equal(tt.expectedErr, err)

			require.NoError(mock.ExpectationsWereMet())
		})
	}
}

func TestUserExistsByName(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name        string
		username    string
		mockSetup   func(sqlmock.Sqlmock)
		expected    bool
		expectedErr error
	}{
		{
			name:     "User Exists",
			username: "existinguser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery("SELECT id FROM `users` WHERE username = \\?;").
					WithArgs("existinguser").
					WillReturnRows(rows)
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name:     "User Does Not Exist",
			username: "nonexistentuser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM `users` WHERE username = \\?;").
					WithArgs("nonexistentuser").
					WillReturnError(sql.ErrNoRows)
			},
			expected:    false,
			expectedErr: nil,
		},
		{
			name:     "Query Error",
			username: "erroruser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM `users` WHERE username = \\?;").
					WithArgs("erroruser").
					WillReturnError(errors.New("query failed"))
			},
			expected:    false,
			expectedErr: errors.New("failed to check if user exists: query failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(err)
			defer db.Close()

			tt.mockSetup(mock)

			exists, err := UserExistsByName(db, tt.username)
			require.Equal(tt.expected, exists)
			require.Equal(tt.expectedErr, err)

			require.NoError(mock.ExpectationsWereMet())
		})
	}
}

func TestGetUserByUsername(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name        string
		username    string
		mockSetup   func(sqlmock.Sqlmock)
		expected    *User
		expectedErr error
	}{
		{
			name:     "User Found",
			username: "existinguser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow(1, "existinguser", "hashedPassword")
				mock.ExpectQuery("SELECT id, username, password_hash FROM `users` WHERE username = \\?;").
					WithArgs("existinguser").
					WillReturnRows(rows)
			},
			expected: &User{
				ID:           1,
				Name:         "existinguser",
				PasswordHash: "hashedPassword",
			},
			expectedErr: nil,
		},
		{
			name:     "User Not Found",
			username: "nonexistentuser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password_hash FROM `users` WHERE username = \\?;").
					WithArgs("nonexistentuser").
					WillReturnError(sql.ErrNoRows)
			},
			expected:    nil,
			expectedErr: nil,
		},
		{
			name:     "Query Error",
			username: "erroruser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password_hash FROM `users` WHERE username = \\?;").
					WithArgs("erroruser").
					WillReturnError(errors.New("query failed"))
			},
			expected:    nil,
			expectedErr: errors.New("failed to get user by name: query failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(err)
			defer db.Close()

			tt.mockSetup(mock)

			user, err := GetUserByUsername(db, tt.username)
			require.Equal(tt.expected, user)
			require.Equal(tt.expectedErr, err)

			require.NoError(mock.ExpectationsWereMet())
		})
	}
}
