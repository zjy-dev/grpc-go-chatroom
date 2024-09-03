package db

import (
	"database/sql"
	"fmt"
)

type User struct {
	ID           int64
	Name         string
	PasswordHash string
}

// InsertUser inserts a new user into the database, and returns the new user's ID.
func InsertUser(db *sql.DB, username, password_hash string) (int64, error) {
	ret, err := db.Exec("INSERT INTO `users` (`username`, `password_hash`) VALUES (?, ?);",
		username, password_hash)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user to database: %v", err)
	}
	return ret.LastInsertId()
}

// UserExistsByName checks if a user exists in the database
func UserExistsByName(db *sql.DB, username string) (bool, error) {
	query := "SELECT id FROM `users` WHERE username = ?;"

	row := db.QueryRow(query, username)

	var id int64
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if user exists: %v", err)
	}

	return true, nil
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	query := "SELECT id, username, password_hash FROM `users` WHERE username = ?;"

	row := db.QueryRow(query, username)

	var id int64
	var uname string
	var password_hash string

	if err := row.Scan(&id, &uname, &password_hash); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by name: %v", err)
	}

	return &User{
		ID:           id,
		Name:         uname,
		PasswordHash: password_hash,
	}, nil
}
