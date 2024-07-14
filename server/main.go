package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/J-Y-Zhang/grpc-go-chatroom/internal/chat"
	"google.golang.org/grpc"
)

type chatServer struct {
	pb.UnimplementedChatServiceServer
	clientsMap map[string]bool
}

func (cs *chatServer) LogIn(ctx context.Context, user *pb.User) (*pb.BaseResp, error) {
	if user.GetUsername() == "" {
		return &pb.BaseResp{Code: 1, Message: "username is empty"}, nil
	}
	cs.clientsMap[user.GetUsername()] = true
	return &pb.BaseResp{Code: 0, Message: ""}, nil
}

func (cs *chatServer) LogOut(ctx context.Context, user *pb.User) (*pb.BaseResp, error) {
	if user.GetUsername() == "" {
		return &pb.BaseResp{Code: 1, Message: "username is empty"}, nil
	}

	if !cs.clientsMap[user.GetUsername()] {
		return &pb.BaseResp{Code: 1, Message: "User has not logged in"}, nil
	}

	delete(cs.clientsMap, user.GetUsername())
	return &pb.BaseResp{Code: 0, Message: ""}, nil
}

func (cs *chatServer) Chat(stream pb.ChatService_ChatServer) error {
	for {
		// Receive message from client
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// TODO: broadcast message to all clients and delete next debugging line
		msg.Text = "Server received: " + msg.Text
		stream.Send(msg)
	}
}

func newServer() *chatServer {
	s := &chatServer{clientsMap: make(map[string]bool)}
	return s
}

func main() {
	// TODO: use dynamic port
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 50051))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterChatServiceServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
