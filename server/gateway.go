package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/zjy-dev/grpc-go-chatroom/api/chat/v1"
)

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			return
		}

		// log.Printf("CORS passed! address: %s, request path: %s, http version: %d, Content-Type: %s",
		// 	r.RemoteAddr, r.URL.Path, r.ProtoMajor, r.Header.Get("Content-Type"))
		next.ServeHTTP(w, r)
	})
}

func gatewayMux() http.Handler {
	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterChatServiceHandlerFromEndpoint(context.Background(), mux, fmt.Sprintf("localhost:%d", port), opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	return cors(mux)
}
