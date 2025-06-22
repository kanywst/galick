## Prometheus Pushgateway Integration

Galick can push load test metrics to a Prometheus Pushgateway, enabling integration with your monitoring stack and alerting systems.

### Configuration

Add Pushgateway settings to your `loadtest.yaml`:

```yaml
report:
  # ...existing report settings...
  pushgateway:
    url: "http://pushgateway.example.com:9091"
    labels:
      instance: galick-load-tests
      environment: production
```

### Command-line options

You can also specify Pushgateway settings via command-line flags:

```bash
galick run --pushgateway-url http://pushgateway.example.com:9091 \
           --push-labels instance=ci-run,build_number=123
```

Or use environment variables:

```bash
GALICK_PUSHGATEWAY_URL=http://pushgateway.example.com:9091 galick run
```

### Metrics

The following metrics are pushed to Prometheus:

* `galick_requests` - Total number of requests executed
* `galick_success_rate` - Success rate (0-1)
* `galick_latency_min_ms` - Minimum latency in milliseconds
* `galick_latency_mean_ms` - Mean latency in milliseconds
* `galick_latency_p50_ms` - 50th percentile (median) latency in milliseconds
* `galick_latency_p90_ms` - 90th percentile latency in milliseconds
* `galick_latency_p95_ms` - 95th percentile latency in milliseconds
* `galick_latency_p99_ms` - 99th percentile latency in milliseconds
* `galick_latency_max_ms` - Maximum latency in milliseconds
* `galick_latency_stddev_ms` - Standard deviation of latency in milliseconds
* `galick_throughput_rps` - Throughput in requests per second
* `galick_duration_seconds` - Test duration in seconds

### Job and Labels

Metrics are pushed with a job name in the format `galick_<environment>_<scenario>`. 

Additional labels include:
* Any labels specified in the configuration or command line
* A `timestamp` label with the Unix timestamp of when the test was completed

### Using with Grafana

Example Prometheus query to visualize P95 latency:

```
galick_latency_p95_ms{job="galick_production_api_endpoints"}
```

Example alert rule for high latency:

```yaml
groups:
- name: galick-alerts
  rules:
  - alert: HighApiLatency
    expr: galick_latency_p95_ms{job="galick_production_api_endpoints"} > 200
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High API latency detected"
      description: "P95 latency is over 200ms for the production API endpoints load test"
```
