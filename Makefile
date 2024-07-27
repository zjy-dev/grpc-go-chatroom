-include .env
.PHONY: proto
proto:
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/proto/chat.proto

.PHONY: clean
clean:
	rm -rf ./bin/*

.PHONY: server
server:
	go run server/auth.go server/main.go
.PHONY: client

name := ""
.PHONY: client
client:
	@go run client/main.go -n=${name}

.PHONY: build
build:
	@mkdir -p ./bin
	@go build -o ./bin/grpc-go-chatroom-server ./server
	@echo "Server built!"
	@go build -o ./bin/grpc-go-chatroom-client ./client 
	@echo "Client built!"
	@echo "Now check ./bin for binaries!"

.PHONY: install
install:
	# TODO

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	# TODO: Make this more graceful
	@go clean -testcache
	go test -cover ./server ./internal/jwt ./internal/tokensource ./internal/middlewares

.PHONY: coverage-html
coverage-html:
	@go clean -testcache && rm -f ./all.coverage.out
	@go test -coverprofile=./all.coverage.out ./...
	@go tool cover -html=./all.coverage.out -o ./coverage.html
	@rm -f ./all.coverage.out
	@echo "Coverage report generated! Open ./coverage.html in your browser to view it." 
