package report

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushMetrics(t *testing.T) {
	// Create a mock Pushgateway server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate the request
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Check that we're pushing to the right endpoint (job name and labels)
		path := r.URL.Path
		// We don't check the exact path because the timestamp will be different each time
		if !strings.Contains(path, "/metrics/job/galick_test") {
			t.Errorf("Unexpected path: %s", path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Success case
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Test with valid parameters
	t.Run("Valid push", func(t *testing.T) {
		err := PushMetrics(
			mockServer.URL,
			"galick_test",
			map[string]string{
				"instance":     "test_instance",
				"build_number": "123",
			},
			map[string]float64{
				"requests":     100,
				"success_rate": 0.995,
				"latency_p95":  150,
			},
		)

		assert.NoError(t, err)
	})

	// Test with empty URL
	t.Run("Empty URL", func(t *testing.T) {
		err := PushMetrics(
			"",
			"galick_test",
			map[string]string{"instance": "test_instance"},
			map[string]float64{"requests": 100},
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pushgateway URL cannot be empty")
	})

	// Test with empty job name
	t.Run("Empty job name", func(t *testing.T) {
		err := PushMetrics(
			mockServer.URL,
			"",
			map[string]string{"instance": "test_instance"},
			map[string]float64{"requests": 100},
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job name cannot be empty")
	})

	// Test with server error
	t.Run("Server error", func(t *testing.T) {
		// Create a server that always returns an error
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Internal server error")
		}))
		defer errorServer.Close()

		err := PushMetrics(
			errorServer.URL,
			"galick_test",
			map[string]string{"instance": "test_instance"},
			map[string]float64{"requests": 100},
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to push metrics to Pushgateway")
	})
}

func TestExtractPrometheusMetrics(t *testing.T) {
	// Test with sample metrics
	metrics := &Metrics{
		SuccessRate: 99.5, // Percentage
		Latencies: LatencyMetrics{
			Min:    1000000,  // 1ms in ns
			Mean:   5000000,  // 5ms in ns
			P50:    4000000,  // 4ms in ns
			P90:    8000000,  // 8ms in ns
			P95:    10000000, // 10ms in ns
			P99:    15000000, // 15ms in ns
			Max:    20000000, // 20ms in ns
			StdDev: 3000000,  // 3ms in ns
		},
		Throughput: 100.5,       // requests per second
		Duration:   30000000000, // 30s in ns
		Requests:   3015,
	}

	promMetrics := ExtractPrometheusMetrics(metrics)

	// Verify that the metrics were correctly converted
	assert.Equal(t, float64(3015), promMetrics["requests"])
	assert.Equal(t, 0.995, promMetrics["success_rate"]) // 99.5% -> 0.995
	assert.Equal(t, 1.0, promMetrics["latency_min_ms"])
	assert.Equal(t, 5.0, promMetrics["latency_mean_ms"])
	assert.Equal(t, 4.0, promMetrics["latency_p50_ms"])
	assert.Equal(t, 8.0, promMetrics["latency_p90_ms"])
	assert.Equal(t, 10.0, promMetrics["latency_p95_ms"])
	assert.Equal(t, 15.0, promMetrics["latency_p99_ms"])
	assert.Equal(t, 20.0, promMetrics["latency_max_ms"])
	assert.Equal(t, 3.0, promMetrics["latency_stddev_ms"])
	assert.Equal(t, 100.5, promMetrics["throughput_rps"])
	assert.Equal(t, 30.0, promMetrics["duration_seconds"])
}
