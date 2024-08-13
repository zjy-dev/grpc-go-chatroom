package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	pb "github.com/zjy-dev/grpc-go-chatroom/api/chat/v1"
	"github.com/zjy-dev/grpc-go-chatroom/internal/tokensource"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	upgrader websocket.Upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type WebSocketServer struct {
	grpcClient pb.ChatServiceClient
}

func mustNewGRPCClient() (*grpc.ClientConn, pb.ChatServiceClient) {
	// Create a new client connection to the server
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	fmt.Println("connected to server")

	// Return the connection and client
	return conn, pb.NewChatServiceClient(conn)
}

func newWebSocketServer() (*WebSocketServer, error) {
	_, client := mustNewGRPCClient()
	return &WebSocketServer{grpcClient: client}, nil
}

func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// get token from request
	token := r.URL.Query().Get("token")
	// log.Printf("address: %s, token: %s\n", r.RemoteAddr, token)

	// Check if the token is empty
	if token == "" {
		log.Printf("token is empty")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Upgrade the request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("failed to upgrade to websocket: %v", err)
		return
	}
	defer func() {
		ws.Close()
	}()

	// Create a stream to the server
	stream, err := s.grpcClient.Chat(context.Background(), grpc.PerRPCCredentials(tokensource.New(token)))
	if err != nil {
		log.Panicf("client.Chat failed: %v\n", err)
	}

	go func() {
		for {
			// Receive message from the server
			msg, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatalf("failed to receive from server: %v", err)
			}

			// return message to the web browser
			ws.WriteJSON(msg)
		}
	}()

	for {
		var req pb.ChatRequest
		err := ws.ReadJSON(&req)
		if err != nil {
			log.Printf("failed to read from websocket: %v", err)
			break
		}
		err = stream.Send(&req)
		if err != nil {
			log.Printf("failed to send to gRPC stream: %v", err)
			break
		}
	}
}

func websocketMux() *http.ServeMux {
	wsServer, err := newWebSocketServer()
	if err != nil {
		log.Fatalf("failed to create WebSocket server: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsServer.handleWebSocket)
	return mux
}
