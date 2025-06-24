# Galick Helm Chart Examples

This directory contains sample files demonstrating various use cases of the Galick Helm Chart.

## File List

- `basic-test.yaml` - Basic load test configuration example
- `scheduled-test.yaml` - Scheduled load test configuration example
- `distributed-test.yaml` - Distributed load test configuration example
- `production-test.yaml` - Production environment configuration example
- `demo-server-test.yaml` - Demo server test example

## Usage

Install Helm Chart using each sample file:

```bash
# Basic test
helm install basic-test ./helm-chart/galick -f docs/examples/basic-test.yaml

# Scheduled test
helm install scheduled-test ./helm-chart/galick -f docs/examples/scheduled-test.yaml

# Distributed test
helm install distributed-test ./helm-chart/galick -f docs/examples/distributed-test.yaml

# Production test
helm install production-test ./helm-chart/galick -f docs/examples/production-test.yaml

# Demo server test
helm install demo-test ./helm-chart/galick -f docs/examples/demo-server-test.yaml
```

## Sample Details

### basic-test.yaml
- **Purpose**: Initial testing and operation verification
- **Features**: Simple test using httpbin.org
- **Duration**: 1 minute
- **Load**: 5 requests/second

### scheduled-test.yaml
- **Purpose**: Periodic monitoring and health checks
- **Features**: Automatic execution every 4 hours via CronJob
- **Duration**: 5 minutes
- **Load**: 2 requests/second

### distributed-test.yaml
- **Purpose**: High-load and large-scale testing
- **Features**: Parallel execution with 3 Pods
- **Duration**: 10 minutes
- **Load**: 300 requests/second (100/s × 3 pods)

### production-test.yaml
- **Purpose**: Performance testing in production environment
- **Features**: Enhanced security, high-performance node usage
- **Duration**: 15 minutes
- **Load**: 50 requests/second

### demo-server-test.yaml
- **Purpose**: Learning and demonstration
- **Features**: Uses built-in demo server
- **Duration**: 2 minutes
- **Load**: 5 requests/second
