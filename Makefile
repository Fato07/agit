.PHONY: build install test clean

BINARY=agit
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/fathindos/agit/cmd.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) .

install:
	go install $(LDFLAGS) .

test:
	go test ./... -v

clean:
	rm -f $(BINARY)
	go clean

fmt:
	go fmt ./...

lint:
	golangci-lint run

tidy:
	go mod tidy
