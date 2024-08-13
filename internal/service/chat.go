package service

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	pb "github.com/zjy-dev/grpc-go-chatroom/api/chat/v1"
	"github.com/zjy-dev/grpc-go-chatroom/internal/db"
	"github.com/zjy-dev/grpc-go-chatroom/internal/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	JWTContextKey = &jwtContext{}
)

type jwtContext struct{}

// chatServiceServer is a struct that implements the chatServiceServer interface.
type chatServiceServer struct {
	pb.UnimplementedChatServiceServer

	clientsMap  map[string]client // username -> client struct
	receiveChan chan *pb.Message  // receive messages from clients, handled by broadcast routine
	mu          sync.Mutex        // mu guards the clientsMap
}

type client struct {
	messageChan chan *pb.Message
}

func NewChatServiceServer() *chatServiceServer {
	server := &chatServiceServer{
		clientsMap:  make(map[string]client),
		receiveChan: make(chan *pb.Message, 1024),
		mu:          sync.Mutex{},
	}
	go server.Broadcast()
	return server
}

// LogIn is a method that implements the LogIn method of the ChatServiceServer interface.
func (cs *chatServiceServer) LogIn(ctx context.Context, req *pb.LogInRequest) (*pb.LogInResponse, error) {
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
	cs.clientsMap[req.GetUsername()] = client{}

	return &pb.LogInResponse{Token: token}, nil
}

// LogOut is a method that implements the LogOut method of the ChatServiceServer interface.
func (cs *chatServiceServer) LogOut(ctx context.Context, _ *pb.LogOutRequest) (*pb.LogOutResponse, error) {
	// Get the username from the context.
	username, ok := ctx.Value(JWTContextKey).(string)
	if !ok || len(username) == 0 {
		return &pb.LogOutResponse{}, status.Errorf(codes.Unauthenticated, "invalid auth token")
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	// Check if the user exists in the clientsMap.
	cli, ok := cs.clientsMap[username]
	if !ok {
		return &pb.LogOutResponse{}, status.Errorf(codes.NotFound, "user: %s not found", username)
	}

	// Remove the user from the clientsMap.
	// NOTE: Close the messageChan of the user. Otherwise goroutine will block forever.
	if cli.messageChan != nil {
		close(cli.messageChan)
	}
	delete(cs.clientsMap, username)
	return &pb.LogOutResponse{}, nil
}

// Chat is a method that implements the Chat method of the ChatServiceServer interface.
func (cs *chatServiceServer) Chat(stream pb.ChatService_ChatServer) error {
	// Get the username from the context.
	username, ok := stream.Context().Value(JWTContextKey).(string)
	if !ok || len(username) == 0 {
		return status.Errorf(codes.Unauthenticated, "invalid auth token")
	}
	cs.mu.Lock()

	// Check if the user exists in the clientsMap.
	if _, ok := cs.clientsMap[username]; !ok {
		cs.mu.Unlock()
		return status.Errorf(codes.NotFound, "user: %s has not logged in, please log in first", username)
	}

	// Add the user(stream) to the clientsMap.
	// TODO
	cliMessageChan := make(chan *pb.Message, 1<<3)
	cs.clientsMap[username] = client{messageChan: cliMessageChan}
	cs.mu.Unlock()

	go func() {
		// TODO
		// cliMessageChan must be closed after logout
		for msg := range cliMessageChan {
			if err := stream.Send(&pb.ChatResponse{Message: msg}); err != nil {
				log.Printf("failed to send message to client: %v\n", err)
			}
		}
	}()

	for {
		// Receive message from client
		req, err := stream.Recv()

		// TODO: Support other type of messages, currently only support text message
		// Check if the request is valid
		reqNotValid := req == nil || req.GetMessage() == nil || (req.GetMessage().GetType() != pb.MessageType_MESSAGE_TYPE_NORMAL)
		if reqNotValid && err != io.EOF {
			cs.mu.Lock()
			if cliMessageChan != nil {
				close(cliMessageChan)
			}
			delete(cs.clientsMap, username)
			cs.mu.Unlock()
			return status.Errorf(codes.InvalidArgument, "empty request or invalid message type")
		}
		if err != nil {
			cs.mu.Lock()
			if cliMessageChan != nil {
				close(cliMessageChan)
			}
			delete(cs.clientsMap, username)
			cs.mu.Unlock()
			if err == io.EOF {
				return nil
			}
			return status.Errorf(codes.Internal, "failed to receive message from client: %v", err)
		}

		// Send message to broadcast routine
		msg := req.GetMessage()
		msg.Timestamp = time.Now().Unix()
		msg.Username = username
		cs.receiveChan <- msg
	}
}

// Broadcast broadcasts messages to all the clients(Fan-out).
// msg from receiveChan already specified timestamp and username if exists
func (cs *chatServiceServer) Broadcast() {
	dbConn := db.MustConnect("127.0.0.1", 3306, "grpc_go_chatroom")
	defer dbConn.Close()
	for msg := range cs.receiveChan {
		id, err := db.InsertMessage(dbConn, 42, msg.Username, msg.TextContent)
		if err != nil || id == 0 {
			log.Printf("failed to insert message: %v\n", err)
			continue
		}
		msg.MessageNumber = uint64(id)
		cs.mu.Lock()

		for username, cli := range cs.clientsMap {
			if username == msg.Username {
				continue
			}
			cli.messageChan <- msg
		}
		cs.mu.Unlock()
	}
}
