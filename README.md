# Galick [![Build Status](https://github.com/kanywst/galick/actions/workflows/galick-ci.yml/badge.svg)](https://github.com/kanywst/galick/actions/workflows/galick-ci.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/kanywst/galick)](https://goreportcard.com/report/github.com/kanywst/galick) [![PkgGoDev](https://pkg.go.dev/badge/github.com/kanywst/galick)](https://pkg.go.dev/github.com/kanywst/galick)

Galick is a versatile HTTP load‑testing wrapper around [Vegeta](https://github.com/tsenart/vegeta), adding centralized configuration, enhanced reporting, and CI/CD integration. Named after Vegeta's iconic attack, Galick brings Saiyan‑level power to your performance workflows.

![Galick Gun](https://static.wikia.nocookie.net/dragonball/images/2/29/Garlic_Gun.png)

## Features

* **Command‑Line & Go Library**: Use as a CLI tool or import as a library
* **Centralized YAML Config**: Define multiple environments and scenarios in one file
* **Enhanced Reports**: HTML, JSON, and Markdown outputs with threshold markers
* **Threshold Enforcement**: Exit non‑zero on SLA breaches for pipelines
* **Pre/Post Hooks**: Run scripts before and after tests
* **Reusable Templates**: Share scenarios across projects
* **Environment Variables**: Override any setting with `GALICK_*`

## Install

### Using Docker

You can use Galick with Docker:

```bash
# Build the Docker image
docker build -t galick:latest .

# Run load tests using Docker
docker run -v $(pwd)/output:/data/output galick:latest run

# Run with Docker Compose (includes demo server)
docker-compose up
```

### Pre‑compiled Binaries

Download the latest release for your platform from the [releases page](https://github.com/kanywst/galick/releases).

### Using Go

```bash
# Vegeta (required)
go install github.com/tsenart/vegeta@latest

go install github.com/kanywst/galick@latest
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/kanywst/galick.git
cd galick

# Build with version information
make build

# The binary will be created at bin/galick
```

## Quick Start

### 1. Initialize config

```bash
galick init
```

Creates a starter `loadtest.yaml` in your current working directory.

### 2. Run a load test

```bash
galick run           # default scenario & environment
galick run heavy     # specify scenario
GALICK_DEFAULT_ENVIRONMENT=staging galick run
```

### 3. Generate reports

```bash
galick report --format html --format markdown
# results in output/<env>/<scenario>/
```

## Configuration (loadtest.yaml)

```yaml
default:
  environment: dev
  scenario: simple
  output_dir: output

environments:
  dev:
    base_url: http://localhost:8080
    headers:
      Content-Type: application/json
  staging:
    base_url: https://staging.example.com
    headers:
      Authorization: Bearer TOKEN

scenarios:
  simple:
    rate: 10/s
    duration: 30s
    targets:
      - GET /api/health
  heavy:
    rate: 50/s
    duration: 60s
    targets:
      - GET /api/products
      - POST /api/orders

report:
  formats:
    - html
    - json
  thresholds:
    p95: 200ms
    success_rate: 99.0

hooks:
  pre: ./scripts/pre-load.sh
  post: ./scripts/post-load.sh
```

Targets use `METHOD /path`, prefixed by `base_url` from the specified environment.

## Commands

### `galick init`

Creates a starter YAML configuration file in your current directory.

```bash
galick init
```

### `galick run`

Execute load test based on your configuration.

```bash
galick run [<scenario>]
```

Flags:

* `--env, -e`: Environment to use (overrides configuration default)
* `--output-dir, -o`: Output directory (overrides configuration default)
* `--config`: Path to config file (default is ./loadtest.yaml)
* `--ci`: Enable CI mode (exit with non-zero code on threshold violations)

### `galick report`

Generate reports from existing test results.

```bash
galick report --results <results-file> [--format html|json|markdown]
```

Flags:

* `--results, -r`: Path to the results file
* `--dir, -d`: Directory to save reports (defaults to same directory as results)
* `--type, -t`: Report type (json, text, markdown, html)

### `galick version`

Display version information.

```bash
galick version
```

## Usage: CI/CD Integration

Galick is designed to integrate smoothly with CI/CD pipelines. Here's an example GitHub Actions workflow:

```yaml
name: Load Test
on: [push, pull_request]
jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - name: Install Vegeta
        run: |
          wget https://github.com/tsenart/vegeta/releases/download/v12.12.0/vegeta_12.12.0_linux_amd64.tar.gz
          tar xzf vegeta_*.tar.gz && sudo mv vegeta /usr/local/bin/
      - name: Install Galick
        run: |
          make build
          sudo mv bin/galick /usr/local/bin/
      - name: Run Load Tests
        run: CI=true galick run
      - name: Upload Reports
        uses: actions/upload-artifact@v3
        with:
          name: load-test-reports
          path: output/
```

## Usage: Distributed Load Testing

For high-scale tests, Galick can be used in a distributed manner, similar to Vegeta. By splitting the load across multiple machines and then aggregating the results, you can achieve much higher request rates.

Example using multiple machines with a shared config:

1. Setup the same configuration on all machines
2. Run with different output directories on each machine
3. Gather the results
4. Generate combined reports

```bash
# On machine 1
galick run heavy --output-dir output/machine1

# On machine 2
galick run heavy --output-dir output/machine2

# Aggregate results (manual step)
```

## Development

### Requirements

- Go 1.24 or higher
- Vegeta v12.12.0 or higher

### Setup Development Environment

```bash
# Install development dependencies
make setup-dev
```

### Run Tests

```bash
make test
```

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md)

## License

See [LICENSE](LICENSE).
