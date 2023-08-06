.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
	golangci-lint ./...
.PHONY:lint

vet: fmt
	go vet ./...
	# shadow ./...
.PHONY:vet

build: vet
	go build .
.PHONY:build
