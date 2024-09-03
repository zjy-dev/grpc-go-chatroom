//go:build integration_test

package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zjy-dev/grpc-go-chatroom/internal/config"
)

func TestMain(m *testing.M) {
	os.Setenv("DBUSER", "root")
	os.Setenv("DBPASS", "123456")

	code := m.Run()

	os.Unsetenv("DBUSER")
	os.Unsetenv("DBPASS")

	os.Exit(code)
}

func TestGetMessagesIntegration(t *testing.T) {
	require := require.New(t)

	dbConn := MustConnect(config.Mysql.User, config.Mysql.Password, config.Mysql.Host, config.Mysql.Port, config.Mysql.DBName)

	res, err := GetMessages(dbConn)
	require.NotEmpty(res)
	require.Nil(err)
}

func TestInsertMessageIntegration(t *testing.T) {
	require := require.New(t)
	dbConn := MustConnect(config.Mysql.User, config.Mysql.Password, config.Mysql.Host, config.Mysql.Port, config.Mysql.DBName)

	t.Cleanup(func() {
		dbConn.Close()
	})

	id, err := InsertMessage(dbConn, 1, "zjy-dev", "hello")
	require.NotZero(id)
	require.Nil(err)
}
