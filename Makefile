-include .env
.PHONY: proto
proto:
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/proto/chat.proto

.PHONY: clean
clean:
	rm -rf internal/chat/chat.pb.go internal/chat/chat_grpc.pb.go

.PHONY: server
server:
	go run server/main.go
.PHONY: client

name := ""
client:
	go run client/main.go -n=${name}

coverage:
	# TODO: Make this more graceful
	@go clean -testcache
	go test -cover ./server ./internal/jwt ./internal/tokensource ./internal/middlewares

coverage-html:
	@go clean -testcache && rm -f ./all.coverage.out
	@go test -coverprofile=./all.coverage.out ./...
	@go tool cover -html=./all.coverage.out -o ./coverage.html
	@rm -f ./all.coverage.out
	@echo "Coverage report generated! Open ./coverage.html in your browser to view it." 