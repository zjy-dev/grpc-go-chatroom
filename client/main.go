package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/J-Y-Zhang/grpc-go-chatroom/internal/chat"
)

func runChat(client pb.ChatServiceClient) {
	msgs := []*pb.Message{
		{Text: "你好"},
		{Text: "瘦瘦耀"},
		{Text: "你好"},
		{Text: "小肥鑫"},
		{Text: "42"},
	}

	stream, _ := client.Chat(context.Background())

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
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// TODO: dynamic server ip:port
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", 50051), opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewChatServiceClient(conn)

	runChat(client)
}
