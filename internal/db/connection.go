package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

// MustConnect initialize and return a new database connection
func MustConnect(host string, port int64, dbName string) *sql.DB {
	// Capture connection properties.
	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	cfg.DBName = dbName
	cfg.Addr = fmt.Sprintf("%s:%d", host, port)

	// Get a database handle.
	dsnStr := cfg.FormatDSN()
	db, err := sql.Open("mysql", dsnStr)

	if err != nil {
		log.Panic(err)
	}

	err = db.Ping()
	if err != nil {
		log.Panic(err)
	}

	return db
}
