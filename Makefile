all: build test benchmark

build:
	go build -v ./...

test:
	go test -v -race -covermode=atomic ./...

benchmark:
	go test -v -bench=. ./...
