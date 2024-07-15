package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/J-Y-Zhang/grpc-go-chatroom/internal/jwt"
	"github.com/J-Y-Zhang/grpc-go-chatroom/internal/middlewares"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/J-Y-Zhang/grpc-go-chatroom/internal/proto"
	authmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
)

type chatServer struct {
	pb.UnimplementedChatServiceServer
	clientsMap map[string]pb.ChatService_ChatServer // username -> client stream
	mu         sync.Mutex
}

func (cs *chatServer) LogIn(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	// Check if username is empty
	if req.GetUsername() == "" {
		return &pb.LoginResp{}, status.Errorf(codes.InvalidArgument, "username is empty")
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Check if user has already logged in
	if _, ok := cs.clientsMap[req.GetUsername()]; ok {
		return &pb.LoginResp{}, status.Errorf(codes.AlreadyExists, "user: %s has already logged in", req.GetUsername())
	}

	// Generate jwt
	token, err := jwt.GenerateJwt(req.GetUsername())
	if err != nil {
		return &pb.LoginResp{}, status.Errorf(codes.Internal, "failed to generate jwt: %v", err)
	}

	// Login(Register) and return jwt
	cs.clientsMap[req.GetUsername()] = nil
	return &pb.LoginResp{Token: token}, nil
}

func (cs *chatServer) LogOut(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	// Get username from context
	username, ok := ctx.Value("username").(string)
	if !ok || len(username) == 0 {
		return &pb.Empty{}, status.Errorf(codes.Unauthenticated, "invalid auth token")
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if _, ok := cs.clientsMap[username]; !ok {
		return &pb.Empty{}, status.Errorf(codes.NotFound, "user: %s not found", username)
	}

	delete(cs.clientsMap, username)
	return &pb.Empty{}, nil
}

func (cs *chatServer) Chat(stream pb.ChatService_ChatServer) error {
	username, ok := stream.Context().Value("username").(string)
	if !ok || len(username) == 0 {
		return status.Errorf(codes.Unauthenticated, "invalid auth token")
	}
	cs.mu.Lock()
	if _, ok := cs.clientsMap[username]; !ok {
		cs.mu.Unlock()
		return status.Errorf(codes.NotFound, "user: %s has not logged in, please log in first", username)
	}
	cs.clientsMap[username] = stream
	cs.mu.Unlock()
	for {
		// Receive message from client
		msg, err := stream.Recv()
		if err == io.EOF {
			cs.mu.Lock()
			delete(cs.clientsMap, username)
			cs.mu.Unlock()
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive message from client: %v", err)
		}

		// TODO: delete or alter the next debugging line
		msg.Text = "Server received: " + msg.Text
		go func(msg *pb.Message) {
			cs.mu.Lock()
			defer cs.mu.Unlock()
			for _, client := range cs.clientsMap {
				if err := client.Send(msg); err != nil {
					log.Printf("failed to send message to client: %v", err)
				}
			}
		}(msg)

	}
}

func newServer() *chatServer {
	s := &chatServer{clientsMap: make(map[string]pb.ChatService_ChatServer)}
	return s
}

func authFunc(ctx context.Context) (context.Context, error) {
	token, err := authmiddleware.AuthFromMD(ctx, "bearer")

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}

	claims, err := jwt.ParseJwt(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}
	subject, err := claims.GetSubject()
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}

	fmt.Println("subject: ", subject)
	// TODO:
	return context.WithValue(ctx, "username", subject), nil
}

func main() {

	// TODO: use dynamic port
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 50051))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(

		grpc.StreamInterceptor(authmiddleware.StreamServerInterceptor(authFunc)),
		grpc.UnaryInterceptor(middlewares.UnaryServerInterceptorWithBypassMethods(authFunc, "LogIn")),
	)
	pb.RegisterChatServiceServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
