.PHONY: build test lint clean install run coverage

BINARY_NAME=veo3
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

build:
	go build -ldflags "-X github.com/jasongoecke/go-veo3/pkg/cli.Version=$(VERSION) -X github.com/jasongoecke/go-veo3/pkg/cli.BuildTime=$(BUILD_TIME)" -o $(BINARY_NAME) ./cmd/veo3

run:
	go run ./cmd/veo3

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total

lint:
	golangci-lint run

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out

install:
	go install ./cmd/veo3
