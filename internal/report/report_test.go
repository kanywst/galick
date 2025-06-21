package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kanywst/galick/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestGenerateReportFormats tests that reports can be generated in different formats
func TestGenerateReportFormats(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock result file (this would normally be output from Vegeta)
	// This is a simplified binary version for testing
	resultFilePath := filepath.Join(tempDir, "results.bin")
	err = os.WriteFile(resultFilePath, []byte("mock-vegeta-data"), 0644)
	assert.NoError(t, err)

	// Create mock metrics for markdown generation
	mockMetrics := &Metrics{
		SuccessRate: 0.995,
		Latencies: LatencyMetrics{
			Min:    10000000,
			Mean:   50000000,
			P50:    45000000,
			P90:    90000000,
			P95:    120000000,
			P99:    180000000,
			Max:    250000000,
			StdDev: 35000000,
		},
		Throughput: 100.5,
		Duration:   30000000000,
		Requests:   3000,
	}

	// Create a mock Reporter that doesn't actually execute vegeta
	mockReporter := &Reporter{
		execCommand: func(_ string, args ...string) ([]byte, error) {
			// Just return sample data based on the format
			if len(args) > 0 {
				switch args[0] {
				case "report":
					if len(args) > 1 && args[1] == "-type=json" {
						return []byte(`{"success":1.0,"latency":{"95th":0.15}}`), nil
					}
					if len(args) > 1 && args[1] == "-type=html" {
						return []byte("<html><body>Report</body></html>"), nil
					}
					return []byte("Success rate: 100%\nLatency p95: 150ms"), nil
				}
			}
			return []byte("Unknown format"), nil
		},
	}

	// Test JSON format
	jsonPath := filepath.Join(tempDir, "report.json")
	err = mockReporter.GenerateReport(resultFilePath, jsonPath, "json", nil)
	assert.NoError(t, err)
	jsonContent, err := os.ReadFile(jsonPath)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonContent), "success")

	// Test HTML format
	htmlPath := filepath.Join(tempDir, "report.html")
	err = mockReporter.GenerateReport(resultFilePath, htmlPath, "html", nil)
	assert.NoError(t, err)
	htmlContent, err := os.ReadFile(htmlPath)
	assert.NoError(t, err)
	assert.Contains(t, string(htmlContent), "<html>")

	// Test Markdown format
	mdPath := filepath.Join(tempDir, "report.md")

	// For markdown, we'll use our markdown generator instead of vegeta's output
	mdReport, err := mockReporter.GenerateMarkdownReport("Test Scenario", "Test Env", mockMetrics, nil)
	assert.NoError(t, err)

	err = os.WriteFile(mdPath, []byte(mdReport), 0644)
	assert.NoError(t, err)

	mdContent, err := os.ReadFile(mdPath)
	assert.NoError(t, err)
	assert.Contains(t, string(mdContent), "# Load Test Report")
}

// TestCheckThresholds tests the threshold validation functionality
func TestCheckThresholds(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		metrics     *Metrics
		thresholds  map[string]string
		expectError bool
	}{
		{
			name: "all metrics within thresholds",
			metrics: &Metrics{
				SuccessRate: 0.995,
				Latencies: LatencyMetrics{
					P95: 150000000, // 150ms in nanoseconds
				},
			},
			thresholds: map[string]string{
				"success_rate": "95.0",
				"p95":          "200ms",
			},
			expectError: false,
		},
		{
			name: "success rate below threshold",
			metrics: &Metrics{
				SuccessRate: 0.90,
				Latencies: LatencyMetrics{
					P95: 150000000, // 150ms in nanoseconds
				},
			},
			thresholds: map[string]string{
				"success_rate": "95.0",
				"p95":          "200ms",
			},
			expectError: true,
		},
		{
			name: "p95 latency above threshold",
			metrics: &Metrics{
				SuccessRate: 0.995,
				Latencies: LatencyMetrics{
					P95: 250000000, // 250ms in nanoseconds
				},
			},
			thresholds: map[string]string{
				"success_rate": "95.0",
				"p95":          "200ms",
			},
			expectError: true,
		},
	}

	// Create a reporter
	reporter := NewReporter()

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := reporter.CheckThresholds(tc.metrics, tc.thresholds)
			if tc.expectError {
				assert.Error(t, err)
				assert.False(t, result)
			} else {
				assert.NoError(t, err)
				assert.True(t, result)
			}
		})
	}
}

// TestGenerateMarkdownReport tests markdown report generation with threshold annotations
func TestGenerateMarkdownReport(t *testing.T) {
	// Create a reporter
	reporter := NewReporter()

	// Create test metrics
	metrics := &Metrics{
		SuccessRate: 0.955,
		Latencies: LatencyMetrics{
			Min:    10000000,
			Mean:   50000000,
			P50:    45000000,
			P90:    90000000,
			P95:    120000000,
			P99:    180000000,
			Max:    250000000,
			StdDev: 35000000,
		},
		Throughput: 100.5,
		Duration:   30000000000,
		Requests:   3000,
	}

	// Define thresholds
	thresholds := map[string]string{
		"success_rate": "95.0",
		"p95":          "100ms", // This will be exceeded
	}

	// Generate report
	report, err := reporter.GenerateMarkdownReport("test-scenario", "test-env", metrics, thresholds)
	assert.NoError(t, err)

	// Check that the report contains expected information
	assert.Contains(t, report, "# Load Test Report")
	assert.Contains(t, report, "**Scenario:** test-scenario")
	assert.Contains(t, report, "**Environment:** test-env")
	assert.Contains(t, report, "Success Rate | 95.5%")
	assert.Contains(t, report, "p95 | 120ms")

	// Check that the threshold warning is included
	assert.Contains(t, report, "Threshold Violations")
	assert.Contains(t, report, "p95: 120ms > 100ms")
}

// TestGenerateReports tests that all report formats can be generated
func TestGenerateReports(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock result file
	resultFilePath := filepath.Join(tempDir, "results.bin")
	err = os.WriteFile(resultFilePath, []byte("mock-vegeta-data"), 0644)
	assert.NoError(t, err)

	// Create a mock config
	cfg := &config.Config{
		Report: config.Report{
			Formats: []string{"json", "html", "markdown"},
			Thresholds: map[string]string{
				"p95":          "100ms",
				"success_rate": "99.0",
			},
		},
	}

	// Create a mock Reporter
	mockReporter := &Reporter{
		execCommand: func(_ string, args ...string) ([]byte, error) {
			// Mock response for report command (used for metrics extraction)
			if args[0] == "report" && args[1] == "-type=json" {
				return []byte(`{
					"success": 0.995,
					"latencies": {
						"min": 10000000,
						"mean": 50000000,
						"50th": 45000000,
						"90th": 90000000,
						"95th": 120000000,
						"99th": 180000000,
						"max": 250000000,
						"stdev": 35000000
					},
					"throughput": 100.5,
					"duration": 30000000000,
					"requests": 3000
				}`), nil
			}
			return []byte("Mocked response"), nil
		},
	}

	// Generate reports
	results, err := mockReporter.GenerateReports(resultFilePath, tempDir, "test-scenario", "test-env", cfg)
	assert.NoError(t, err)

	// Check that we have 3 report files
	assert.Len(t, results, 3)

	// Check that all formats were generated
	formats := make(map[string]bool)
	for _, result := range results {
		formats[result.Format] = true
		assert.FileExists(t, result.FilePath)
	}

	assert.True(t, formats["json"])
	assert.True(t, formats["html"])
	assert.True(t, formats["markdown"])
}
