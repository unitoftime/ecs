all: build test benchmark

build:
	go build -v ./...

test:
#	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go test -v -coverprofile=coverage.out -covermode=count ./...

benchmark:
	go test -v -bench=. ./...

coverage:
	go tool cover -html=coverage.out

bench2:
	go test -run=nothingplease -bench=BenchmarkAddEntity -benchmem -memprofile mem.pprof -cpuprofile cpu.pprof
