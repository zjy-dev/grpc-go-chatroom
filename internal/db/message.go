package db

import (
	"database/sql"
	"fmt"

	pb "github.com/zjy-dev/grpc-go-chatroom/api/chat/v1"
)

func InsertMessage(db *sql.DB, userID int, username, message string) (id int64, err error) {
	ret, err := db.Exec("INSERT INTO chat_message (user_id, username, message) VALUES (?, ?, ?);",
		userID, username, message)
	if err != nil {
		return 0, fmt.Errorf("failed to insert to database: %v", err)
	}

	id, err = ret.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last inserted message ID: %v", err)
	}
	return id, nil
}

func GetMessages(db *sql.DB) ([]*pb.Message, error) {
	query := "SELECT id, user_id, username, message, created_at FROM chat_message;"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %v", err)
	}
	defer rows.Close()

	res := make([]*pb.Message, 0, 32)
	for rows.Next() {
		var id, userID int
		var username, message string
		var createdAt string
		err = rows.Scan(&id, &userID, &username, &message, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		res = append(res, &pb.Message{
			TextContent:   message,
			Username:      username,
			MessageNumber: uint64(id),
		})

	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
