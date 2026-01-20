# Load Testing Tools Analysis

- [Load Testing Tools Analysis](#load-testing-tools-analysis)
  - [1. Vegeta](#1-vegeta)
    - [Implementation](#implementation)
    - [Strengths](#strengths)
    - [Weaknesses](#weaknesses)
  - [2. Hey](#2-hey)
    - [Implementation](#implementation-1)
    - [Strengths](#strengths-1)
    - [Weaknesses](#weaknesses-1)
  - [3. Wrk](#3-wrk)
    - [Implementation](#implementation-2)
    - [Strengths](#strengths-2)
    - [Weaknesses](#weaknesses-2)
  - [4. K6](#4-k6)
    - [Implementation](#implementation-3)
    - [Strengths](#strengths-3)
    - [Weaknesses](#weaknesses-3)
  - [5. Ghz](#5-ghz)
    - [Implementation](#implementation-4)
    - [Strengths](#strengths-4)
    - [Weaknesses](#weaknesses-4)
  - [6. Fortio](#6-fortio)
    - [Implementation](#implementation-5)
    - [Strengths](#strengths-5)
- [The Vision for `galick`](#the-vision-for-galick)
    - [Architecture Goals](#architecture-goals)
    - [Implementation Strategy (Destructive Rewrite)](#implementation-strategy-destructive-rewrite)

This document analyzes the implementation details, strengths, and weaknesses of several popular open-source load testing tools. This analysis serves as the foundation for the architecture of `galick`.

## 1. Vegeta

**Language:** Go
**Core Component:** `lib/attack.go`

### Implementation

- **Concurrency:** Uses a fixed worker pool of goroutines.
- **Pacing:** Uses a `Pacer` interface to calculate sleep times between requests. It employs a centralized "tick" mechanism where a main loop fires signals to workers.
- **Metrics:** Uses `tdigest` (or HdrHistogram concepts) for accurate quantile calculation. Metrics are aggregated in memory.

### Strengths

- **Precision:** The `Pacer` model allows for very precise constant-rate throughput (e.g., exactly 50 QPS).
- **Simplicity:** The library design is clean and easy to embed in other Go programs.
- **Output:** Supports multiple output formats (JSON, Histogram, Text) and can pipe results to plotting tools.

### Weaknesses

- **Coordination Bottleneck:** The centralized pacing channel can theoretically become a contention point at massive concurrency (though rarely an issue in practice).
- **Limited Scripting:** Primarily designed for static URL/Body attacks. Dynamic behavior requires writing Go code or complex piping.

## 2. Hey

**Language:** Go
**Core Component:** `requester/requester.go`

### Implementation

- **Concurrency:** Launches `C` concurrent workers.
- **Pacing:** "Distributed Pacing". Each worker manages its own `time.Ticker`.
- **Metrics:** Collects raw result objects in a buffered channel. A separate goroutine aggregates them.

### Strengths

- **Lightweight:** Very small binary and simple codebase.
- **Easy of Use:** Mimics `ab` (Apache Bench) flags, making it familiar to many.

### Weaknesses

- **Memory Usage:** Storing individual result objects can lead to high memory usage on long-running tests with high throughput.
- **Drift:** Independent tickers in workers can drift, leading to less precise aggregate QPS compared to Vegeta.

## 3. Wrk

**Language:** C
**Core Component:** `src/wrk.c`, `src/net.c`

### Implementation

- **Architecture:** Event-driven using the `ae` event loop (from Redis).
- **Concurrency:** Multithreaded, with each thread running an independent event loop (non-blocking I/O).
- **Scripting:** Embeds Lua (LuaJIT) to allow request generation and response handling.
- **HTTP Parsing:** Uses the high-performance `http-parser` library.

### Strengths

- **Performance:** Best-in-class throughput and low latency overhead due to C + Event Loop architecture.
- **Efficiency:** Minimal CPU/Memory footprint.

### Weaknesses

- **Complexity:** C codebase is harder to maintain and extend.
- **Feature Set:** Limited out-of-the-box features (no HTTP/2, gRPC without plugins).
- **Coordinated Omission:** Standard reporting often suffers from coordinated omission (measuring service time vs response time), though `wrk2` addresses this.

## 4. K6

**Language:** Go (with JavaScript runtime)
**Core Component:** `lib/runner.go`, `lib/executor/`

### Implementation

- **Architecture:** "Virtual User" (VU) model. Each VU is a goroutine running a JavaScript VM (`goja`).
- **Executors:** Modular system to control execution flow (e.g., `ConstantVUs`, `RampingArrivalRate`).
- **Scripting:** Full ES6+ (via Babel/bundling) scripting environment.

### Strengths

- **flexibility:** Extremely scriptable. Can simulate complex user flows (login -> browse -> checkout).
- **UX:** Excellent CLI UX, progress bars, and cloud integration.
- **Ecosystem:** Huge library of extensions (xk6).

### Weaknesses

- **Resource Heavy:** Running a JS VM per VU is CPU/Memory intensive compared to compiled Go or C.
- **Go-JS Bridge:** Marshaling data between Go and JS incurs overhead.

## 5. Ghz

**Language:** Go
**Core Component:** `runner/run.go`

### Implementation

- **Focus:** Specialized for gRPC.
- **Architecture:** Similar to `hey`/`vegeta` but tailored for HTTP/2 and Protobuf.
- **Features:** Supports proto reflection and descriptor sets.

### Strengths

- **gRPC Native:** Handles binary protocols and gRPC specifics natively.

### Weaknesses

- **Niche:** Only for gRPC.

## 6. Fortio

**Language:** Go
**Core Component:** `periodic/periodic.go`

### Implementation

- **Pacing:** Closed-loop feedback system. It calculates "Target Elapsed Time" and sleeps `Target - Actual`.
- **Catch-up:** Explicitly handles "falling behind" by skipping sleeps or requests to maintain average QPS.

### Strengths

- **Reliability:** Very robust statistical model.
- **Server Mode:** Includes a web UI and echo server.

---

# The Vision for `galick`

`galick` will synthesize these learnings into a next-generation tool.

### Architecture Goals

1. **Engine:** Go-based with a modular "Attacker" interface.
   *  *Mode 1 (Throughput):* Precise, tick-based pacing (Vegeta-style) for stress testing APIs.
   *  *Mode 2 (Flow):* VU-based execution (K6-style) for simulating user journeys.
2. **Scripting:** Embedded **Starlark** (Python dialect). It offers the flexibility of K6's JS but with much lower overhead and safer execution in Go.
3. **Performance:** Non-blocking I/O patterns where possible, minimizing allocations in the hot path.
4. **UX:** Beautiful TUI (Bubbletea) for real-time feedback.
5. **Protocols:** Pluggable architecture supporting HTTP/1.1, HTTP/2, and gRPC out of the box.

### Implementation Strategy (Destructive Rewrite)

We will completely replace the current repository content with this new architecture.
