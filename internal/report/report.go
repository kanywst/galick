// Package report provides functionality for generating load test reports
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/kanywst/galick/internal/config"
)

// Reporter generates reports from Vegeta output files
type Reporter struct {
	execCommand func(cmd string, args ...string) ([]byte, error)
}

// ReportResult contains information about a generated report
type ReportResult struct {
	FilePath string
	Format   string
	Passed   bool
}

// Metrics represents the load test metrics
type Metrics struct {
	SuccessRate float64       `json:"success"`
	Latencies   LatencyMetrics `json:"latencies"`
	Throughput  float64       `json:"throughput"`
	Duration    int64         `json:"duration"`
	Requests    int           `json:"requests"`
}

// LatencyMetrics represents latency metrics
type LatencyMetrics struct {
	Min    int64   `json:"min"`
	Mean   int64   `json:"mean"`
	P50    int64   `json:"50th"`
	P90    int64   `json:"90th"`
	P95    int64   `json:"95th"`
	P99    int64   `json:"99th"`
	Max    int64   `json:"max"`
	StdDev int64   `json:"stdev"`
}

// NewReporter creates a new reporter instance
func NewReporter() *Reporter {
	return &Reporter{
		execCommand: func(cmd string, args ...string) ([]byte, error) {
			return exec.Command(cmd, args...).CombinedOutput()
		},
	}
}

// GenerateReports generates all configured report formats
func (r *Reporter) GenerateReports(resultFile, outputDir, scenario, environment string, cfg *config.Config) ([]ReportResult, error) {
	// Extract metrics for threshold validation
	metrics, err := r.extractMetrics(resultFile)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metrics: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Check which formats to generate
	formats := cfg.Report.Formats
	if len(formats) == 0 {
		// Default to text format if none specified
		formats = []string{"text"}
	}

	// Generate each format
	results := make([]ReportResult, 0, len(formats))
	for _, format := range formats {
		outputFile := filepath.Join(outputDir, fmt.Sprintf("report.%s", formatExtension(format)))
		
		err := r.GenerateReport(resultFile, outputFile, format, metrics)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s report: %w", format, err)
		}

		// Check thresholds
		passed := true
		if len(cfg.Report.Thresholds) > 0 {
			passed, _ = r.CheckThresholds(metrics, cfg.Report.Thresholds)
		}

		results = append(results, ReportResult{
			FilePath: outputFile,
			Format:   format,
			Passed:   passed,
		})
	}

	// If markdown is one of the formats, add threshold annotations
	for _, result := range results {
		if result.Format == "markdown" {
			// Generate markdown with threshold annotations
			mdReport, err := r.GenerateMarkdownReport(scenario, environment, metrics, cfg.Report.Thresholds)
			if err != nil {
				return nil, fmt.Errorf("failed to generate markdown report: %w", err)
			}

			// Write the markdown report
			if err := os.WriteFile(result.FilePath, []byte(mdReport), 0644); err != nil {
				return nil, fmt.Errorf("failed to write markdown report: %w", err)
			}
		}
	}

	return results, nil
}

// GenerateReport generates a report in the specified format
func (r *Reporter) GenerateReport(resultFile, outputFile, format string, metrics *Metrics) error {
	var args []string
	var outputData []byte
	var err error

	// For markdown format, we handle it specially
	if format == "markdown" && metrics != nil {
		// Generate basic markdown (will be enhanced later in GenerateReports)
		mdReport, err := r.GenerateMarkdownReport("", "", metrics, nil)
		if err != nil {
			return err
		}
		outputData = []byte(mdReport)
	} else {
		// For other formats, use vegeta's report command
		args = []string{"report"}
		
		// Add format flag for non-text formats
		if format != "text" {
			args = append(args, "-type="+format)
		}
		
		// Add input file
		args = append(args, "-inputs="+resultFile)
		
		// Execute vegeta report command
		outputData, err = r.execCommand("vegeta", args...)
		if err != nil {
			return fmt.Errorf("vegeta report command failed: %w", err)
		}
	}

	// Write the report to file
	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

// CheckThresholds validates metrics against configured thresholds
func (r *Reporter) CheckThresholds(metrics *Metrics, thresholds map[string]string) (bool, error) {
	if metrics == nil || len(thresholds) == 0 {
		return true, nil
	}

	violations := make([]string, 0)

	// Check success rate
	if successRateThreshold, exists := thresholds["success_rate"]; exists {
		threshold, err := strconv.ParseFloat(successRateThreshold, 64)
		if err != nil {
			return false, fmt.Errorf("invalid success_rate threshold value: %s", successRateThreshold)
		}

		if metrics.SuccessRate*100 < threshold {
			violations = append(violations, 
				fmt.Sprintf("success_rate: %.2f%% < %.2f%%", metrics.SuccessRate*100, threshold))
		}
	}

	// Check latency thresholds
	for metric, thresholdStr := range thresholds {
		var latencyNs int64
		var actualValue float64

		switch metric {
		case "p50", "50th":
			latencyNs = metrics.Latencies.P50
			actualValue = float64(latencyNs) / 1000000 // ns to ms
		case "p90", "90th":
			latencyNs = metrics.Latencies.P90
			actualValue = float64(latencyNs) / 1000000
		case "p95", "95th":
			latencyNs = metrics.Latencies.P95
			actualValue = float64(latencyNs) / 1000000
		case "p99", "99th":
			latencyNs = metrics.Latencies.P99
			actualValue = float64(latencyNs) / 1000000
		case "mean":
			latencyNs = metrics.Latencies.Mean
			actualValue = float64(latencyNs) / 1000000
		case "max":
			latencyNs = metrics.Latencies.Max
			actualValue = float64(latencyNs) / 1000000
		case "success_rate", "error_rate":
			// Already handled or will be handled separately
			continue
		default:
			// Skip unknown metrics
			continue
		}

		// Parse threshold value (convert from "200ms" to milliseconds)
		thresholdValue, err := parseLatencyThreshold(thresholdStr)
		if err != nil {
			return false, fmt.Errorf("invalid threshold for %s: %s", metric, err)
		}

		// Compare actual value to threshold
		if actualValue > thresholdValue {
			violations = append(violations, 
				fmt.Sprintf("%s: %.0fms > %.0fms", metric, actualValue, thresholdValue))
		}
	}

	// Check error rate threshold
	if errorRateThreshold, exists := thresholds["error_rate"]; exists {
		threshold, err := strconv.ParseFloat(errorRateThreshold, 64)
		if err != nil {
			return false, fmt.Errorf("invalid error_rate threshold value: %s", errorRateThreshold)
		}

		errorRate := (1 - metrics.SuccessRate) * 100
		if errorRate > threshold {
			violations = append(violations, 
				fmt.Sprintf("error_rate: %.2f%% > %.2f%%", errorRate, threshold))
		}
	}

	if len(violations) > 0 {
		return false, fmt.Errorf("threshold violations: %s", strings.Join(violations, ", "))
	}

	return true, nil
}

// GenerateMarkdownReport generates a markdown report with threshold annotations
func (r *Reporter) GenerateMarkdownReport(scenario, environment string, metrics *Metrics, thresholds map[string]string) (string, error) {
	// Define the markdown template
	const mdTemplate = `# Load Test Report

{{ if .Scenario }}**Scenario:** {{ .Scenario }}{{ end }}
{{ if .Environment }}**Environment:** {{ .Environment }}{{ end }}
**Date:** {{ .Date }}

## Summary

| Metric | Value |
|--------|-------|
| Success Rate | {{ .SuccessRate }}% |
| Requests | {{ .Requests }} |
| Throughput | {{ .Throughput }} req/s |
| Duration | {{ .Duration }}s |

## Latency

| Percentile | Value |
|------------|-------|
| Min | {{ .Min }}ms |
| Mean | {{ .Mean }}ms |
| p50 | {{ .P50 }}ms |
| p90 | {{ .P90 }}ms |
| p95 | {{ .P95 }}ms |
| p99 | {{ .P99 }}ms |
| Max | {{ .Max }}ms |
| Std Dev | {{ .StdDev }}ms |

{{ if .Violations }}
## Threshold Violations ⚠️

{{ range .Violations }}
- ⚠️ Threshold exceeded: {{ . }}
{{ end }}
{{ end }}
`

	// Check thresholds and collect violations
	var violations []string
	if thresholds != nil && len(thresholds) > 0 {
		_, err := r.CheckThresholds(metrics, thresholds)
		if err != nil {
			// Extract the violations from the error message
			msg := err.Error()
			if strings.HasPrefix(msg, "threshold violations: ") {
				violationsStr := strings.TrimPrefix(msg, "threshold violations: ")
				violations = strings.Split(violationsStr, ", ")
			}
		}
	}

	// Create template data
	data := struct {
		Scenario    string
		Environment string
		Date        string
		SuccessRate string
		Requests    int
		Throughput  string
		Duration    string
		Min         string
		Mean        string
		P50         string
		P90         string
		P95         string
		P99         string
		Max         string
		StdDev      string
		Violations  []string
	}{
		Scenario:    scenario,
		Environment: environment,
		Date:        time.Now().Format("2006-01-02 15:04:05"),
		SuccessRate: fmt.Sprintf("%.1f", metrics.SuccessRate*100),
		Requests:    metrics.Requests,
		Throughput:  fmt.Sprintf("%.1f", metrics.Throughput),
		Duration:    fmt.Sprintf("%.1f", float64(metrics.Duration)/1000000000),
		Min:         fmt.Sprintf("%.0f", float64(metrics.Latencies.Min)/1000000),
		Mean:        fmt.Sprintf("%.0f", float64(metrics.Latencies.Mean)/1000000),
		P50:         fmt.Sprintf("%.0f", float64(metrics.Latencies.P50)/1000000),
		P90:         fmt.Sprintf("%.0f", float64(metrics.Latencies.P90)/1000000),
		P95:         fmt.Sprintf("%.0f", float64(metrics.Latencies.P95)/1000000),
		P99:         fmt.Sprintf("%.0f", float64(metrics.Latencies.P99)/1000000),
		Max:         fmt.Sprintf("%.0f", float64(metrics.Latencies.Max)/1000000),
		StdDev:      fmt.Sprintf("%.0f", float64(metrics.Latencies.StdDev)/1000000),
		Violations:  violations,
	}

	// Parse the template
	tmpl, err := template.New("markdown").Parse(mdTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse markdown template: %w", err)
	}

	// Generate the report
	var report strings.Builder
	err = tmpl.Execute(&report, data)
	if err != nil {
		return "", fmt.Errorf("failed to generate markdown report: %w", err)
	}

	return report.String(), nil
}

// extractMetrics extracts metrics from a Vegeta result file
func (r *Reporter) extractMetrics(resultFile string) (*Metrics, error) {
	// Run vegeta report with JSON output
	output, err := r.execCommand("vegeta", "report", "-type=json", "-inputs="+resultFile)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metrics: %w", err)
	}

	// Parse the JSON
	var metrics Metrics
	if err := json.Unmarshal(output, &metrics); err != nil {
		return nil, fmt.Errorf("failed to parse metrics JSON: %w", err)
	}

	return &metrics, nil
}

// formatExtension returns the file extension for a given format
func formatExtension(format string) string {
	switch format {
	case "json":
		return "json"
	case "html":
		return "html"
	case "markdown", "md":
		return "md"
	default:
		return "txt"
	}
}

// parseLatencyThreshold parses a latency threshold string (e.g., "200ms") into milliseconds
func parseLatencyThreshold(threshold string) (float64, error) {
	// Remove whitespace
	threshold = strings.TrimSpace(threshold)

	// If the threshold doesn't have a unit, assume milliseconds
	if _, err := strconv.ParseFloat(threshold, 64); err == nil {
		value, _ := strconv.ParseFloat(threshold, 64)
		return value, nil
	}

	// Check for known units
	if strings.HasSuffix(threshold, "ms") {
		value, err := strconv.ParseFloat(strings.TrimSuffix(threshold, "ms"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid value for milliseconds: %s", threshold)
		}
		return value, nil
	}

	if strings.HasSuffix(threshold, "s") {
		value, err := strconv.ParseFloat(strings.TrimSuffix(threshold, "s"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid value for seconds: %s", threshold)
		}
		return value * 1000, nil // Convert seconds to milliseconds
	}

	return 0, fmt.Errorf("unknown latency unit in threshold: %s", threshold)
}
