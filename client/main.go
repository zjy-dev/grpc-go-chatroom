package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/J-Y-Zhang/grpc-go-chatroom/internal/proto"
	"github.com/J-Y-Zhang/grpc-go-chatroom/internal/tokensource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func runChat(token string, client pb.ChatServiceClient) {
	msgs := []*pb.Message{
		{Text: "你好"},
		{Text: "瘦瘦耀"},
		{Text: "你好"},
		{Text: "小肥鑫"},
		{Text: "42"},
	}

	stream, err := client.Chat(context.Background(), grpc.PerRPCCredentials(tokensource.New(token)))
	if err != nil {
		log.Fatalf("client.Chat failed: %v", err)
	}

	waitc := make(chan struct{})
	go func() {
		// receive chat messages from the server
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("client.RouteChat failed: %v", err)
			}
			log.Printf("Got message %s at %v", in.GetText(), time.Unix(in.GetTimestamp(), 0))
		}
	}()

	for _, msg := range msgs {
		msg.Timestamp = time.Now().Local().Unix()
		err := stream.Send(msg)
		if err != nil {
			log.Fatalf("%v.Send(%v) = %v", client, msg, err)
		}
		time.Sleep(time.Microsecond * 700)
	}

	stream.CloseSend()
	<-waitc

}

func main() {
	// Set up the credentials for the connection.
	// perRPC := oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
	// 	AccessToken: "zjy",
	// })}

	var opts []grpc.DialOption = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// TODO: dynamic server ip:port
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", 50051), opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewChatServiceClient(conn)

	// log in
	loginResp, err := client.LogIn(context.Background(), &pb.LoginReq{
		Username: "zjy",
	})
	if err != nil {
		log.Fatalf("client.LogIn failed: %v", err)
	}

	if loginResp == nil {
		log.Fatalf("server returned an empty response: %v", err)
	}
	if len(loginResp.GetToken()) == 0 {
		log.Fatalf("server returned an empty token: %v", err)
	}

	// chat
	runChat(loginResp.GetToken(), client)

}
