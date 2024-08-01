-include .env

BINART_PATH	:= ./bin
PROTO_PATH := ./internal/proto
SERVER_PATH = ./server
.PHONY: deps
deps:
	go mod tidy
	cd ${PROTO_PATH} && buf dep update


.PHONY: proto
proto:
	cd ${PROTO_PATH} && buf generate

.PHONY: clean
clean:
	rm -rf ${BINART_PATH}/*
	rm -rf ${PROTO_PATH}/*.pb*go

.PHONY: run-server
SERVER_SOURCES := $(shell find ${SERVER_PATH} -type f ! -name '*_test.go')
run-server:
	go run ${SERVER_SOURCES}
.PHONY: client

name := ""
.PHONY: run-client
run-client:
	@go run client/main.go -n=${name}

.PHONY: build
build:
	@mkdir -p ./bin
	@CGO_ENABLED=0 go build -o ./bin/grpc-go-chatroom-server ./server
	@echo "Server built!"
	@CGO_ENABLED=0 go build -o ./bin/grpc-go-chatroom-client ./client 
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

.PHONY: docker
docker:
	@docker build -t yysfg/grpc-go-chatroom .

.PHONY: docker-compose
docker-compose:
	@docker compose up --build
