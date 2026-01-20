BINARY_NAME=galick
GOLANGCI_LINT_VERSION ?= v1.64.6

# Versioning information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(DATE) -s -w"

all: lint test build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) cmd/galick/main.go

test:
	go test -v -race ./...

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f demo.gif

run:
	go run $(LDFLAGS) cmd/galick/main.go

# Tooling
install-tools:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

lint:
	golangci-lint run ./...

audit:
	govulncheck ./...

vhstape:
	vhs demo.tape