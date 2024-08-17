//go:build integration_test

package db

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

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
