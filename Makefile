.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	internal/chat/chat.proto

.PHONY: clean
clean:
	rm -rf internal/chat/chat.pb.go internal/chat/chat_grpc.pb.go

.PHONY: server
server:
	go run server/main.go
.PHONY: client
client:
	go run client/main.go