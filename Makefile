.PHONY: build test clean install release help

BINARY_NAME=slash
VERSION=1.0.0
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

help:
	@echo "Slash Build Targets"
	@echo "==================="
	@echo "  make build          Build slash binary"
	@echo "  make test           Run all tests"
	@echo "  make test-adapters  Run adapter tests only"
	@echo "  make coverage       Generate test coverage report"
	@echo "  make install        Build and install to /usr/local/bin"
	@echo "  make release        Build releases for all platforms"
	@echo "  make clean          Remove build artifacts"

build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	CGO_ENABLED=1 go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/slash

test:
	@echo "Running tests..."
	go test -v -race -timeout 30s ./...

test-adapters:
	@echo "Running adapter tests..."
	go test -v ./internal/adapters/...

coverage:
	@echo "Generating coverage report..."
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

install: build
	@echo "Installing to /usr/local/bin..."
	sudo cp $(BINARY_NAME) /usr/local/bin/

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME) slash.exe
	rm -rf dist/ build/
	rm -f coverage.txt coverage.html
	go clean

release:
	@echo "Building releases for v$(VERSION)..."
	mkdir -p dist/

	# Linux x86_64
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/slash-linux-amd64 ./cmd/slash
	cd dist && tar czf slash_$(VERSION)_linux_amd64.tar.gz slash-linux-amd64

	# macOS x86_64
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/slash-darwin-amd64 ./cmd/slash
	cd dist && tar czf slash_$(VERSION)_darwin_amd64.tar.gz slash-darwin-amd64

	# macOS ARM64 (Apple Silicon)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/slash-darwin-arm64 ./cmd/slash
	cd dist && tar czf slash_$(VERSION)_darwin_arm64.tar.gz slash-darwin-arm64

	# Windows
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/slash-windows-amd64.exe ./cmd/slash
	cd dist && zip slash_$(VERSION)_windows_amd64.zip slash-windows-amd64.exe

	@echo "Releases built in dist/"
	@ls -lh dist/

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	golangci-lint run ./...

vet:
	@echo "Running go vet..."
	go vet ./...
