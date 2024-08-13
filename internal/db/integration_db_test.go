//go:build integration_test

package db

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Setenv("DBUSER", "root")
	os.Setenv("DBPASS", "123456")

	code := m.Run()

	os.Unsetenv("DBUSER")
	os.Unsetenv("DBPASS")

	os.Exit(code)
}

func TestMustConnect(t *testing.T) {
	require := require.New(t)
	require.NotPanics(func() {
		dbConn := MustConnect("127.0.0.1", 3306, "grpc_go_chatroom")
		if dbConn == nil {
			log.Panic("MustConnect should not return nil")
		}
		dbConn.Close()
	})
}

func TestGetMessages(t *testing.T) {
	tmpDir := t.TempDir()
	p := filepath.Join(tmpDir, "get_messages.sql")
	// TODO: Figure out os.Mode*
	os.WriteFile(p, []byte("SELECT id, user_id, username, message, created_at FROM chat_messages;"), os.ModePerm)
	require := require.New(t)

	dbConn := MustConnect("127.0.0.1", 3306, "grpc_go_chatroom")

	res, err := GetMessages(dbConn, tmpDir)
	require.NotEmpty(res)
	require.Nil(err)
}

func TestInsertMessage(t *testing.T) {
	require := require.New(t)
	dbConn := MustConnect("127.0.0.1", 3306, "grpc_go_chatroom")

	t.Cleanup(func() {
		dbConn.Close()
	})

	id, err := InsertMessage(dbConn, 1, "zjy-dev", "hello")
	require.NotZero(id)
	require.Nil(err)
}
