// Package report provides functionality for generating load test reports.
package report

import (
	"fmt"
	"time"

	"github.com/kanywst/galick/internal/constants"
	gerrors "github.com/kanywst/galick/internal/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// PushMetrics sends metrics to a Prometheus Pushgateway instance.
// It takes a Pushgateway URL, job name, labels, and metrics to push.
// Returns an error if the push operation fails.
func PushMetrics(pushURL string, jobName string, labels map[string]string, metrics map[string]float64) error {
	if pushURL == "" {
		return gerrors.ErrPushgatewayURLEmpty
	}

	if jobName == "" {
		return gerrors.ErrJobNameEmpty
	}

	// Create a new registry for our metrics
	registry := prometheus.NewRegistry()

	// Create gauges for each metric
	gauges := make(map[string]prometheus.Gauge)
	for name := range metrics {
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("galick_%s", name),
			Help: fmt.Sprintf("Galick load test metric: %s", name),
		})
		gauges[name] = gauge
		registry.MustRegister(gauge)
	}

	// Set gauge values
	for name, value := range metrics {
		gauges[name].Set(value)
	}

	// Create a pusher for the registry
	pusher := push.New(pushURL, jobName).
		Gatherer(registry).
		Grouping("timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	// Add custom labels if provided
	for k, v := range labels {
		pusher = pusher.Grouping(k, v)
	}

	// Push metrics to the Pushgateway
	if err := pusher.Push(); err != nil {
		return fmt.Errorf("failed to push metrics to Pushgateway: %w", err)
	}

	return nil
}

// ExtractPrometheusMetrics converts a Metrics struct to a map of metric name to value
// suitable for sending to Prometheus.
func ExtractPrometheusMetrics(metrics *Metrics) map[string]float64 {
	// Convert success rate from percentage to decimal
	successRate := metrics.SuccessRate
	if successRate > 1 {
		successRate /= 100
	}

	// Convert nanosecond latencies to milliseconds for better readability
	return map[string]float64{
		"requests":          float64(metrics.Requests),
		"success_rate":      successRate,
		"latency_min_ms":    float64(metrics.Latencies.Min) / constants.NanoToMillisecond,
		"latency_mean_ms":   float64(metrics.Latencies.Mean) / constants.NanoToMillisecond,
		"latency_p50_ms":    float64(metrics.Latencies.P50) / constants.NanoToMillisecond,
		"latency_p90_ms":    float64(metrics.Latencies.P90) / constants.NanoToMillisecond,
		"latency_p95_ms":    float64(metrics.Latencies.P95) / constants.NanoToMillisecond,
		"latency_p99_ms":    float64(metrics.Latencies.P99) / constants.NanoToMillisecond,
		"latency_max_ms":    float64(metrics.Latencies.Max) / constants.NanoToMillisecond,
		"latency_stddev_ms": float64(metrics.Latencies.StdDev) / constants.NanoToMillisecond,
		"throughput_rps":    metrics.Throughput,
		"duration_seconds":  float64(metrics.Duration) / constants.NanoToSecond,
	}
}
