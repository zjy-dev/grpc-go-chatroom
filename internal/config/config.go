package config

import (
	"embed"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

var (
	//go:embed config.yaml
	configFS embed.FS

	Mysql  *mysqlConfig
	Server *serverConfig
	JWT    *jwtConfig
)

type mysqlConfig struct {
	Host     string
	Port     uint64
	User     string
	Password string
	DBName   string
}

type serverConfig struct {
	Port uint64
}

type jwtConfig struct {
	JWTKey string
}

func init() {
	loadConfigs()
}

func loadConfigs() {

	content, err := configFS.ReadFile("config.yaml")
	if err != nil || len(content) == 0 {
		log.Fatalf("failed to read embed config file: %v", err)
	}

	config := viper.New()
	config.SetConfigFile("config.yaml")
	config.SetConfigType("yaml")

	configFile, err := configFS.Open("config.yaml")
	if err != nil {
		log.Fatalf("failed to open embed config file: %v", err)
	}

	err = config.ReadConfig(configFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	// Mysql
	dbPortStr := os.Getenv("GRPC_GO_CHATROOM_DBPORT")
	if dbPortStr == "" {
		log.Fatal("dbport is not set or invalid, check environment variables")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("failed to parse dbport: %v", err)
	}
	mysqlPass := ""
	if t := os.Getenv("GRPC_GO_CHATROOM_DBPASS"); t != "" {
		mysqlPass = t
	} else {
		f, err := os.Open("/run/secrets/db-password")
		if err != nil {
			log.Fatalf("failed to open dbpass secret file: %v", err)
		}
		dbPassRaw, err := io.ReadAll(f)
		dbPass := strings.TrimSpace(string(dbPassRaw))
		if err != nil || dbPass == "" {
			log.Fatalf("failed to read dbpass secret file: %v", err)
		}
		mysqlPass = string(dbPass)
	}

	Mysql = &mysqlConfig{
		Host:     os.Getenv("GRPC_GO_CHATROOM_DBHOST"),
		Port:     uint64(dbPort),
		User:     os.Getenv("GRPC_GO_CHATROOM_DBUSER"),
		DBName:   os.Getenv("GRPC_GO_CHATROOM_DBNAME"),
		Password: mysqlPass,
	}

	if Mysql.Host == "" || Mysql.Port <= 0 || Mysql.DBName == "" || Mysql.User == "" || Mysql.Password == "" {
		log.Fatal("invalid mysql config, check env vars")
	}

	// Server
	Server = &serverConfig{
		Port: uint64(config.GetInt("server.port")),
	}

	if Server.Port <= 0 {
		log.Fatalf("invalid server config, check config.yaml")
	}

	// JWT
	if t := os.Getenv("GRPC_GO_CHATROOM_JWT_KEY"); t != "" {
		JWT = &jwtConfig{
			JWTKey: t,
		}
	} else {
		f, err := os.Open("/run/secrets/jwy-key")
		if err != nil {
			log.Fatalf("failed to open jwt-key secret file: %v", err)
		}
		jwtKeyRaw, err := io.ReadAll(f)
		jwtKey := strings.TrimSpace(string(jwtKeyRaw))
		if err != nil || jwtKey == "" {
			log.Fatalf("failed to read jwt-key secret file: %v", err)
		}
		JWT = &jwtConfig{
			JWTKey: string(jwtKey),
		}
	}
}
