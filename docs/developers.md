# Developer Guide

This document provides comprehensive information for developers who want to contribute to the Galick project or customize it for their own use.

## Project Structure

```
galick/
├── cmd/                 # Command line interfaces
│   ├── cli/             # Main entry point
│   └── galick/          # Command implementations
├── internal/            # Internal packages
│   ├── config/          # Configuration handling
│   ├── hooks/           # Pre/post execution hooks
│   ├── report/          # Report generation
│   │   └── templates/   # Report templates (HTML, etc.)
│   └── runner/          # Load test execution
├── scripts/             # Helper scripts
│   ├── demo-server.go   # Demo HTTP server
│   ├── pre-load.sh      # Example pre-hook
│   ├── post-load.sh     # Example post-hook
│   └── run-demo.sh      # Demo runner
├── docs/                # Documentation
├── examples/            # Example configurations
└── bin/                 # Build outputs (git-ignored)
```

## Development Setup

1. Clone the repository
   ```bash
   git clone https://github.com/kanywst/galick.git
   cd galick
   ```

2. Ensure Go 1.18+ is installed
   ```bash
   go version
   ```

3. Install Vegeta:
   ```bash
   go install github.com/tsenart/vegeta@latest
   ```

4. Build the project:
   ```bash
   go build -o bin/galick ./cmd/cli
   ```

5. Run the demo server:
   ```bash
   go run scripts/demo-server.go
   ```

6. In another terminal, run Galick:
   ```bash
   ./bin/galick run
   ```

## Running Tests

```bash
go test ./...
```

For tests with coverage:

```bash
go test -cover ./...
```

## Running the Demo

We provide a demo script that:
1. Starts a local HTTP server
2. Runs the load test against it
3. Generates a demo GIF using VHS

Prerequisites:
- [VHS](https://github.com/charmbracelet/vhs) installed
- Go 1.18+ installed
- Vegeta installed

```bash
# Make the scripts executable
chmod +x scripts/*.sh

# Run the demo
./scripts/run-demo.sh
```

## VHS Demo Tape

Here's a sample [VHS](https://github.com/charmbracelet/vhs) tape to create a demo GIF of Galick:

```bash
# Galick Demo
Output galick-demo.gif

Set Shell zsh
Set FontSize 18
Set Width 1200
Set Height 600
Set Padding 20

# Initialize a new config
Type "galick init"
Enter
Sleep 2s

# Show the generated config
Type "cat loadtest.yaml | head -n 20"
Enter
Sleep 3s

# Run a load test
Type "galick run simple"
Enter
Sleep 5s

# Show reports
Type "cat output/dev/simple/report.md"
Enter
Sleep 5s

# Generate a different report format
Type "galick report simple --format json"
Enter
Sleep 2s

# Show the JSON report
Type "cat output/dev/simple/report.json | jq"
Enter
Sleep 3s
```

## Adding New Report Formats

Galick's report system is designed to be extensible. To add a new report format:

1. Create a new file in the `internal/report` package
2. Implement the required interfaces
3. Register your new format in the report package

Example for adding a new report format:

```go
package report

// NewXMLReporter creates a new XML reporter
func NewXMLReporter() Reporter {
    return &XMLReporter{}
}

// XMLReporter implements the Reporter interface for XML format
type XMLReporter struct{}

// Generate creates an XML report from the given results
func (r *XMLReporter) Generate(results *vegeta.Results, config *config.Config) (string, error) {
    // Implementation goes here
    return xmlContent, nil
}

// Register your new format in the init() function of report.go
func init() {
    RegisterReporter("xml", NewXMLReporter)
}
```

## Contributing Guidelines

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin feature/your-feature-name`
5. Submit a pull request

Please ensure your code follows our style guidelines and includes appropriate tests.

## Code Style

We follow standard Go style guidelines:

- Run `go fmt` before committing
- Use `golint` and `go vet` to check your code
- Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines

## Release Process

1. Update version in `cmd/cli/main.go`
2. Update CHANGELOG.md
3. Create a new tag: `git tag v1.0.0`
4. Push the tag: `git push origin v1.0.0`
5. GitHub Actions will automatically build and publish the release
