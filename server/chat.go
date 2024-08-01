package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/zjy-dev/grpc-go-chatroom/internal/jwt"
	pb "github.com/zjy-dev/grpc-go-chatroom/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// chatServiceServer is a struct that implements the ChatServiceServer interface.
type chatServiceServer struct {
	pb.UnimplementedChatServiceServer // UnimplementedChatServiceServer is the server API for ChatService service.

	clientsMap map[string]pb.ChatService_ChatServer // username -> client stream
	mu         sync.Mutex                           // mu guards the clientsMap
}

func newChatServer() *chatServiceServer {
	return &chatServiceServer{clientsMap: make(map[string]pb.ChatService_ChatServer)}
}

// LogIn is a method that implements the LogIn method of the ChatServiceServer interface.
func (cs *chatServiceServer) LogIn(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	log.Printf("LogIn: %s\n", req.GetUsername())
	if req.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is empty")
	}

	// Lock the mutex to prevent concurrent access to the clientsMap.
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Check if the user has already logged in.
	if _, ok := cs.clientsMap[req.GetUsername()]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "user: %s has already logged in", req.GetUsername())
	}

	// Generate a JWT token for the user.
	token, err := jwt.GenerateJwt(req.GetUsername())
	if err != nil {

		return nil, status.Errorf(codes.Internal, "failed to generate jwt: %v", err)
	}

	// Add the user to the clientsMap.
	cs.clientsMap[req.GetUsername()] = nil

	return &pb.LoginResp{Token: token}, nil
}

// LogOut is a method that implements the LogOut method of the ChatServiceServer interface.
func (cs *chatServiceServer) LogOut(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {

	// Get the username from the context.
	username, ok := ctx.Value("username").(string)
	if !ok || len(username) == 0 {
		return &pb.Empty{}, status.Errorf(codes.Unauthenticated, "invalid auth token")
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	// Check if the user exists in the clientsMap.
	if _, ok := cs.clientsMap[username]; !ok {
		return &pb.Empty{}, status.Errorf(codes.NotFound, "user: %s not found", username)
	}

	// Remove the user from the clientsMap.
	delete(cs.clientsMap, username)
	return &pb.Empty{}, nil
}

// Chat is a method that implements the Chat method of the ChatServiceServer interface.
func (cs *chatServiceServer) Chat(stream pb.ChatService_ChatServer) error {
	// Get the username from the context.
	username, ok := stream.Context().Value("username").(string)
	if !ok || len(username) == 0 {
		return status.Errorf(codes.Unauthenticated, "invalid auth token")
	}
	cs.mu.Lock()
	// Check if the user exists in the clientsMap.
	if _, ok := cs.clientsMap[username]; !ok {
		cs.mu.Unlock()
		return status.Errorf(codes.NotFound, "user: %s has not logged in, please log in first", username)
	}
	// Add the user to the clientsMap.

	cs.clientsMap[username] = stream

	cs.mu.Unlock()

	for {
		// Receive message from client
		msg, err := stream.Recv()
		if err != io.EOF && msg == nil {
			cs.mu.Lock()
			delete(cs.clientsMap, username)
			cs.mu.Unlock()
			return status.Errorf(codes.InvalidArgument, "empty message")
		}
		if err != nil {
			cs.mu.Lock()
			delete(cs.clientsMap, username)
			cs.mu.Unlock()
			if err == io.EOF {
				return nil
			}
			return status.Errorf(codes.Internal, "failed to receive message from client: %v", err)
		}

		// Send message to all clients
		// TODO: CHECK TIMESTAMP
		newMsg := &pb.Message{Text: fmt.Sprintf("%s: %s", username, msg.Text), Timestamp: msg.GetTimestamp()}
		cs.mu.Lock()
		for _, client := range cs.clientsMap {
			go func() {
				if err := client.Send(newMsg); err != nil {
					log.Printf("failed to send message to client: %v", err)
				}
			}()
		}
		cs.mu.Unlock()
	}
}
