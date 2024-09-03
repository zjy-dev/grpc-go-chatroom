package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/zjy-dev/grpc-go-chatroom/internal/config"
	"github.com/zjy-dev/grpc-go-chatroom/internal/tokensource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/zjy-dev/grpc-go-chatroom/api/chat/v1"
)

var (
	username string
	token    string
)

// mustLogin function logs in the user to the chatroom
func mustLogin(client pb.ChatServiceClient) {

	// Send a login request to the server
	loginResp, err := client.LogInOrRegister(context.Background(), &pb.LogInOrRegisterRequest{
		Username: username,
	})
	if err != nil {
		log.Fatalf("client.LogIn failed: %v", err)
	}

	// Check if the server returned an empty response
	if loginResp == nil {
		log.Fatalf("server returned an empty client.LogIn response: %v", err)
	}
	// Check if the server returned an empty token
	if len(loginResp.GetToken()) == 0 {
		log.Fatalf("server returned an empty token: %v", err)
	}

	// Set the token
	token = loginResp.GetToken()
}

// chat
func chat(client pb.ChatServiceClient) {
	if len(token) == 0 {
		log.Panicln("no token found")
	}

	// Create a stream to the server
	stream, err := client.Chat(context.Background(), grpc.PerRPCCredentials(tokensource.New(token)))
	if err != nil {
		log.Panicf("client.Chat failed: %v\n", err)
	}
	// Create a channel to wait for the server to send a message
	waitc := make(chan struct{})

	// Start a goroutine to receive messages from the server
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("failed to receive from server: %v", err)
			}

			// Print the message from the server
			fmt.Printf("[%s] %s %s\n", time.Unix(resp.GetMessage().GetTimestamp(), 0).Format("2006-01-02 15:04:05"),
				resp.GetMessage().GetUsername(), resp.GetMessage().GetTextContent())
		}
	}()

	// Create a scanner to read from standard input
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		// Create a message to send to the server
		msg := &pb.Message{
			Type:        pb.MessageType_MESSAGE_TYPE_NORMAL,
			TextContent: scanner.Text(),
		}

		// Send the request to the server
		err := stream.Send(&pb.ChatRequest{Message: msg})
		if err != nil {
			log.Fatalf("%v.Send(%v) = %v", client, msg, err)
		}
	}

	// Check if there was an error reading from standard input
	if err := scanner.Err(); err != nil {
		stream.CloseSend()
		log.Fatalf("reading standard input: %v", err)
	}
	stream.CloseSend()

	// Wait for the server to send a message
	<-waitc
}

// mustNewClient function creates a new client connection to the server
func mustNewClient() (*grpc.ClientConn, pb.ChatServiceClient) {

	// Create a new client connection to the server
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", config.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	fmt.Println("connecting to server")

	// Return the connection and client
	return conn, pb.NewChatServiceClient(conn)
}

// main function is the entry point of the program
func main() {
	// Create a new cli app
	chatroomClient := &cli.App{
		Name:  "gRPC-go-chatroom client",
		Usage: "gRPC-go chatroom client, written for learning purposes",

		// Define the action to be taken when the app is run
		Action: func(cCtx *cli.Context) error {

			// Create a new client connection to the server
			conn, client := mustNewClient()
			defer conn.Close()

			// Log in the user to the chatroom
			mustLogin(client)

			fmt.Printf("Hello, %s! Welcome to the chatroom!\n", username)
			fmt.Println("Input your message and hit enter to shoot it, and havvvve a nice chat!")
			// Run the chatroom
			chat(client)
			return nil
		},
		// Define the flags for the app
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:        "port",
				Aliases:     []string{"p"},
				Value:       config.Server.Port,
				Usage:       "the server port",
				Destination: nil,
			},

			&cli.StringFlag{
				Name:        "name",
				Aliases:     []string{"n"},
				Required:    true,
				Usage:       "username for the chatroom",
				Destination: &username,
			},
		},
	}

	// Run the cli app
	if err := chatroomClient.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
