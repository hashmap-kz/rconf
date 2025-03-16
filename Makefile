# Variables
SOURCES := $(shell find . -name '*.go')
BINARY := rconf
COV_REPORT := coverage.txt
TEST_FLAGS := -v -race -timeout 30s
INSTALL_DIR := /usr/local/bin

# Default target
.PHONY: all
all: build

# Build the binary (GOARCH=amd64 GOOS=linux; -o $(BINARY))
.PHONY: build
build: $(SOURCES)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/ main.go

# Install the binary to /usr/local/bin
.PHONY: install
install: build
	@echo "Installing bin/$(BINARY) to $(INSTALL_DIR)..."
	@install -m 0755 bin/$(BINARY) $(INSTALL_DIR)

# Run unit tests
.PHONY: test
test:
	go test $(TEST_FLAGS) ./...

# Run tests with coverage
.PHONY: test-cov
test-cov:
	go test -coverprofile=$(COV_REPORT) ./...
	go tool cover -html=$(COV_REPORT)

# Setup docker-compose (for running integration tests in a sandbox)
.PHONY: compose-setup
compose-setup:
	docker compose -f integration/docker-compose.yml up --build -d

# Cleanup docker-compose
.PHONY: compose-teardown
compose-teardown:
	docker compose -f integration/docker-compose.yml down

# Run integration tests
.PHONY: test-integration
test-integration: compose-setup
	RCONF_INTEGRATION_TESTS_AVAILABLE=0xcafebabe go test -v internal/runner/*.go
	$(MAKE) compose-teardown

# Lint the code
.PHONY: lint
lint:
	golangci-lint run ./...

# Check goreleaser
.PHONY: snapshot
snapshot:
	goreleaser release --skip sign --skip publish --snapshot --clean

# Format the code
.PHONY: format
format:
	go fmt ./...

# Clean build artifacts
.PHONY: clean
clean:
	@rm -rf bin/ $(COV_REPORT)
