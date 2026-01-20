# Galick

<div align="center">
  <img src="demo.gif" width="100%" alt="Galick Demo" />
</div>

<div align="center">

[![Go Report Card](https://goreportcard.com/badge/github.com/kanywst/galick)](https://goreportcard.com/report/github.com/kanywst/galick)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kanywst/galick)](https://github.com/kanywst/galick)
[![CI](https://github.com/kanywst/galick/actions/workflows/ci.yml/badge.svg)](https://github.com/kanywst/galick/actions/workflows/ci.yml)

**The next-generation load testing tool for modern engineering teams.**

</div>

---

**Galick** is a high-performance, extensible load testing tool written in Go. It synthesizes the best features of tools like *Vegeta*, *k6*, and *wrk* into a single, cohesive binary.

## Why Galick?

Most load testing tools force a trade-off:

* **Vegeta** is fast and precise but **static** (can't easily generate random data).
* **K6** is scriptable but **heavy** (JavaScript VM per user uses lots of RAM).
* **Wrk** is fast but **complex** (requires Lua and is hard to extend).

**Galick solves this by combining:**

1. **Starlark Scripting:** Write dynamic scenarios in Python-like syntax without the overhead of a full JS VM.
2. **Go Performance:** Uses Go's lightweight concurrency for high throughput.
3. **Precise Pacing:** Distributed constant-rate pacing for accurate stress testing.

## Installation

### From Source

```bash
go install github.com/kanywst/galick/cmd/galick@latest
```

### Docker

```bash
docker pull ghcr.io/kanywst/galick:latest
```

You can run Galick directly using Docker:

```bash
# Basic usage
docker run --rm ghcr.io/kanywst/galick:latest --url https://example.com --qps 10 --duration 10s --headless

# Running a local script
docker run --rm -v $(pwd)/attack.star:/attack.star ghcr.io/kanywst/galick:latest --script /attack.star --headless
```

## Quick Start

### 1. Static Mode (Like Vegeta)

Perfect for hitting a single endpoint with constant load.

```bash
galick --url https://api.example.com/v1/users \
       --qps 50 \
       --workers 10 \
       --duration 30s
```

### 2. Dynamic Scripting Mode (Like K6)

Perfect for generating random data, dynamic paths, or complex payloads.

Create a file named `attack.star`:

```python
# attack.star
def request():
    return {
        "method": "POST",
        "url": "https://httpbin.org/post",
        "body": '{"user_id": 123, "timestamp": "now"}'
    }
```

Run it:

```bash
galick --script attack.star --qps 100 --duration 1m
```

### 3. Docker / CI Mode (Headless)

For CI/CD pipelines or background jobs, use `--headless` to disable the TUI and output clean logs.

```bash
galick --url https://api.example.com --headless
```

Or using Docker Compose:

```bash
docker-compose run --rm galick
```

## Options

|     Flag     | Shorthand | Default |                 Description                 |
| :----------: | :-------: | :-----: | :-----------------------------------------: |
|   `--url`    |   `-u`    |    -    |        Target URL (for static mode)         |
|  `--script`  |   `-s`    |    -    | Path to Starlark script (for dynamic mode)  |
|  `--method`  |   `-m`    |  `GET`  |                 HTTP Method                 |
|   `--qps`    |   `-q`    |  `50`   |             Queries Per Second              |
| `--workers`  |   `-w`    |  `10`   |        Number of concurrent workers         |
| `--duration` |   `-d`    |  `10s`  |            Duration of the test             |
| `--headless` |           | `false` | Run without TUI (recommended for CI/Docker) |

## Architecture

Galick is built on a modular engine:

1. **Attacker Interface:** Pluggable protocols (HTTP, Starlark Script).
2. **Engine:** Manages concurrency and pacing.
3. **Metrics:** Thread-safe, HdrHistogram for accurate P99 latency.
4. **Reporter:** Beautiful TUI powered by Bubbletea.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.
