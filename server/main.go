package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	authmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/spf13/viper"
	pb "github.com/zjy-dev/grpc-go-chatroom/api/chat/v1"
	"github.com/zjy-dev/grpc-go-chatroom/internal/middlewares"
	"github.com/zjy-dev/grpc-go-chatroom/internal/service"
	"github.com/zjy-dev/grpc-go-chatroom/internal/utils"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

var (
	port      int64
	mysqlHost string
	mysqlPort int64
	dbName    string
)

func main() {
	loadConfigs()
	// Serve websocket & gRPC-gateway
	mux := websocketMux()
	mux.Handle("/", gatewayMux())

	// Serve frontend
	mux.Handle("/static", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	grpcServer := grpcServer()

	if port == 0 {
		log.Fatalf("server.port is not set or invalid, check config.yaml")
	}
	log.Printf("server will listen at 0.0.0.0:%d", port)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), combinedProtocolHandler(grpcServer, mux)))
}

func combinedProtocolHandler(grpcServer *grpc.Server, gatewayAndWebsocketMux *http.ServeMux) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("address: %s, request path: %s, http version: %d, Content-Type: %s",
		// 	r.RemoteAddr, r.URL.Path, r.ProtoMajor, r.Header.Get("Content-Type"))
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			gatewayAndWebsocketMux.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func grpcServer() *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.ChainStreamInterceptor(authmiddleware.StreamServerInterceptor(authFunc)),
		// Exclude the "LogIn" method from authentication.
		grpc.UnaryInterceptor(middlewares.UnaryServerAuthInterceptorWithBypassMethods(authFunc, "LogIn")),
	)

	pb.RegisterChatServiceServer(grpcServer, service.NewChatServiceServer())

	return grpcServer
}

func loadConfigs() {
	utils.MustLoadEnvFile()
	config := viper.New()

	config.AddConfigPath(".")
	config.AddConfigPath("..")
	config.AddConfigPath("./config")
	config.AddConfigPath("../config")
	config.SetConfigName("config")
	config.SetConfigType("yaml")

	if err := config.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	port = int64(config.GetInt("server.port"))
	if port == 0 {
		log.Fatalf("server.port is not set or invalid, check config.yaml")
	}
}
