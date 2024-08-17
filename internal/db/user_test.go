//go:build unit_test

package db

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestInsertUser(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name         string
		username     string
		passwordHash string
		mockBehavior func(mock sqlmock.Sqlmock)
		expectedID   int64
		expectErr    bool
	}{
		{
			name:         "Successful Insert",
			username:     "testuser",
			passwordHash: "hashedpassword",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO `user` \\(`username`, `password_hash`\\) VALUES \\(\\?, \\?\\);").
					WithArgs("testuser", "hashedpassword").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedID: 1,
			expectErr:  false,
		},
		{
			name:         "Insert Failure",
			username:     "testuser",
			passwordHash: "hashedpassword",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO `user` \\(`username`, `password_hash`\\) VALUES \\(\\?, \\?\\);").
					WithArgs("testuser", "hashedpassword").
					WillReturnError(errors.New("insert failed"))
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
			id, err := InsertUser(db, tt.username, tt.passwordHash)

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
