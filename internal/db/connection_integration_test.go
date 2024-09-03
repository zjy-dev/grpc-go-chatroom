//go:build integration_test

package db

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zjy-dev/grpc-go-chatroom/internal/config"
)

func TestMustConnect(t *testing.T) {
	require := require.New(t)
	require.NotPanics(func() {
		dbConn := MustConnect(config.Mysql.User, config.Mysql.Password, config.Mysql.Host, config.Mysql.Port, config.Mysql.DBName)
		if dbConn == nil {
			log.Panic("MustConnect should not return nil")
		}
		dbConn.Close()
	})
}
