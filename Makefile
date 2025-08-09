include .env
export

.PHONY: build_rest
build_rest:
	go build -v ./cmd/rest

.PHONY: build_grpc
build_grpc:
	go build -v ./cmd/grpc

.PHONY: test
test:
	go test -v -race -timeout 30s ./...

.PHONY: proto
proto:
	protoc -I="C:/Program Files/protoc/include"  -I protos/proto \
		--go_out=./protos/gen/go --go_opt paths=source_relative \
		--go-grpc_out=./protos/gen/go --go-grpc_opt paths=source_relative \
		./protos/proto/*/*.proto

# .DEFAULT_GOAL := build