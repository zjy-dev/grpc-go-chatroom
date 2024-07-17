package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/urfave/cli/v2"
	"github.com/zjy-dev/grpc-go-chatroom/internal/jwt"
	"github.com/zjy-dev/grpc-go-chatroom/internal/middlewares"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	pb "github.com/zjy-dev/grpc-go-chatroom/internal/proto"
)

var port int64

// chatServer is a struct that implements the ChatServiceServer interface.
type chatServer struct {
	pb.UnimplementedChatServiceServer // UnimplementedChatServiceServer is the server API for ChatService service.

	clientsMap map[string]pb.ChatService_ChatServer // username -> client stream
	mu         sync.Mutex                           // mu guards the clientsMap
}

// LogIn is a method that implements the LogIn method of the ChatServiceServer interface.
func (cs *chatServer) LogIn(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {

	// Check if the username is empty.
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
func (cs *chatServer) LogOut(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {

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
func (cs *chatServer) Chat(stream pb.ChatService_ChatServer) error {
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

			cs.mu.Unlock()

		}
	}
}

// newChatServer creates a new chatServer instance.
func newChatServer() *chatServer {
	s := &chatServer{clientsMap: make(map[string]pb.ChatService_ChatServer)}
	return s
}

func main() {
	// Create a new CLI app.
	chatroomServer := &cli.App{
		Name:  "grpc-go-chatroom server",                                // Set the name of the app.
		Usage: "grpc-go chatroom server, written for learning purposes", // Set the usage message.

		Action: func(cCtx *cli.Context) error {
			// Listen for incoming connections on the specified port.
			lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}

			// Create a new gRPC server.
			grpcServer := grpc.NewServer(

				// Set the stream interceptor to authenticate incoming stream requests.
				grpc.StreamInterceptor(authmiddleware.StreamServerInterceptor(authFunc)),
				// Set the unary interceptor to authenticate incoming unary requests.
				// Exclude the "LogIn" method from authentication.
				grpc.UnaryInterceptor(middlewares.UnaryServerAuthInterceptorWithBypassMethods(authFunc, "LogIn")),
			)

			// Register the chat service server with the gRPC server.
			pb.RegisterChatServiceServer(grpcServer, newChatServer())

			// Start the server and serve the chat requests.
			grpcServer.Serve(lis)
			return nil
		},

		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:        "port",            // Set the name of the flag.
				Aliases:     []string{"p"},     // Set the aliases of the flag.
				Value:       50051,             // Set the default value of the flag.
				Usage:       "the server port", // Set the usage message of the flag.
				Destination: &port,             // Set the destination of the flag.
			},
		},
	}

	// Run the CLI app and handle any errors that occur.
	if err := chatroomServer.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
