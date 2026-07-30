.PHONY: build build-release deps fmt vet test clean help

VERSION ?= dev
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"
BINARY := gopro-validator

help:
	@echo "gopro-validator"
	@echo ""
	@echo "Development:"
	@echo "  make deps           - Download and tidy Go dependencies"
	@echo "  make build          - Build binary (VERSION=dev)"
	@echo "  make build-release  - Build optimized release binary"
	@echo "  make fmt            - Format Go source files"
	@echo "  make vet            - Run go vet static analysis"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Remove build artifacts"

deps:
	go mod download
	go mod tidy

build:
	@echo "Building gopro-validator (version: $(VERSION))..."
	go build $(LDFLAGS) -o $(BINARY) ./cmd/

build-release:
	@echo "Building optimized release binary (version: $(VERSION))..."
	go build \
		-ldflags "-s -w -X main.Version=$(VERSION)" \
		-trimpath \
		-o $(BINARY) ./cmd/

fmt:
	gofmt -w cmd/ internal/

vet:
	go vet ./...

test:
	go test ./...

clean:
	rm -f $(BINARY)
