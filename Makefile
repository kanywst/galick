# version information
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
VERSION=$(shell git describe --tags --exact-match --always 2>/dev/null || echo "dev")
DATE=$(shell date +'%FT%TZ')

# dependencies
GO_VERSION=1.24
VEGETA_VERSION=v12.12.0
GOLANGCI_LINT_VERSION=v2.1.6

.PHONY: build clean test lint setup-dev

# build target
build:
	@echo "Building Galick version $(VERSION) (commit: $(COMMIT), built: $(DATE))"
	@mkdir -p bin
	CGO_ENABLED=0 go build -v -a \
	-ldflags '-s -w -X github.com/kanywst/galick/internal/cli.version=$(VERSION) \
	-X github.com/kanywst/galick/internal/cli.commit=$(COMMIT) \
	-X github.com/kanywst/galick/internal/cli.buildDate=$(DATE)' \
	-o bin/galick ./cmd/galick

# run tests
test:
	go test -v -race ./...

# run linters
lint:
	go vet ./...
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

# clean up
clean:
	rm -rf bin/
	rm -f galick

# install the binary to GOPATH/bin
install: build
	cp bin/galick $(GOPATH)/bin/galick

# build and run the demo server
demo-server:
	go build -o demo-server ./scripts/demo-server.go
	./demo-server

# setup development environment
setup-dev:
	@echo "Setting up development environment..."
	@echo "Installing Go $(GO_VERSION)..."
	@echo "Please install Go $(GO_VERSION) manually from https://golang.org/dl/"
	@echo "Installing Vegeta $(VEGETA_VERSION)..."
	@if [ "$(shell uname)" = "Darwin" ]; then \
		brew install vegeta; \
	else \
		echo "Please install Vegeta $(VEGETA_VERSION) manually from https://github.com/tsenart/vegeta/releases"; \
	fi
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@if [ "$(shell uname)" = "Darwin" ]; then \
		brew install golangci-lint; \
	else \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi
	@echo "Development environment setup complete!"

# HELP command
help:
	@echo "Galick Makefile Help"
	@echo "--------------------"
	@echo "make build       - Build the galick binary"
	@echo "make test        - Run all tests"
	@echo "make lint        - Run linters"
	@echo "make clean       - Remove build artifacts"
	@echo "make install     - Install galick to GOPATH/bin"
	@echo "make demo-server - Build and run the demo server"
	@echo "make setup-dev   - Set up development environment"
	@echo ""
	@echo "Dependencies:"
	@echo "  Go:            $(GO_VERSION)"
	@echo "  Vegeta:        $(VEGETA_VERSION)"
	@echo "  golangci-lint: $(GOLANGCI_LINT_VERSION)"

# default target
default: build
