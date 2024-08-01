package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/viper"
	"github.com/zjy-dev/grpc-go-chatroom/internal/middlewares"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	authmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	pb "github.com/zjy-dev/grpc-go-chatroom/internal/proto"
)

var (
	port int64
)

func main() {
	mux := websocketMux()
	mux.Handle("/", gatewayMux())
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
		grpc.StreamInterceptor(authmiddleware.StreamServerInterceptor(authFunc)),
		// Exclude the "LogIn" method from authentication.
		grpc.UnaryInterceptor(middlewares.UnaryServerAuthInterceptorWithBypassMethods(authFunc, "LogIn")),
	)

	pb.RegisterChatServiceServer(grpcServer, newChatServer())

	return grpcServer
}

func init() {
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
