// Package report provides functionality for generating load test reports.
package report

import (
	// embed is needed for HTML template embedding.
	_ "embed"
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
	gerrors "github.com/kanywst/galick/internal/errors"
)

// Format constants.
const (
	FormatHTML     = "html"
	FormatMarkdown = "markdown"
	FormatJSON     = "json"
	FormatText     = "text"
)

// HTML template is provided in html.go.

// Reporter generates reports from Vegeta output files.
type Reporter struct {
	execCommand func(cmd string, args ...string) ([]byte, error)
}

// Result contains information about a generated report.
type Result struct {
	FilePath string
	Format   string
	Passed   bool
}

// Metrics represents the load test metrics.
type Metrics struct {
	SuccessRate float64        `json:"success"`
	Latencies   LatencyMetrics `json:"latencies"`
	Throughput  float64        `json:"throughput"`
	Duration    int64          `json:"duration"`
	Requests    int            `json:"requests"`
}

// LatencyMetrics represents latency metrics.
type LatencyMetrics struct {
	Min    int64 `json:"min"`
	Mean   int64 `json:"mean"`
	P50    int64 `json:"50th"`
	P90    int64 `json:"90th"`
	P95    int64 `json:"95th"`
	P99    int64 `json:"99th"`
	Max    int64 `json:"max"`
	StdDev int64 `json:"stdev"`
}

// NewReporter creates a new reporter instance.
func NewReporter() *Reporter {
	return &Reporter{
		execCommand: func(cmd string, args ...string) ([]byte, error) {
			return exec.Command(cmd, args...).CombinedOutput()
		},
	}
}

// GenerateReports generates all configured report formats.
func (r *Reporter) GenerateReports(
	resultFile, outputDir, scenario, environment string,
	cfg *config.Config,
) ([]Result, error) {
	// Validate input parameters
	if err := r.validateReportInputs(resultFile, outputDir); err != nil {
		return nil, err
	}

	// Extract metrics for threshold validation
	metrics, err := r.extractMetrics(resultFile)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metrics: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine which formats to generate
	formats := r.getReportFormats(cfg)

	// Generate each format
	results, err := r.generateBasicReports(resultFile, outputDir, formats, metrics, cfg)
	if err != nil {
		return nil, err
	}

	// Enhance rich format reports (markdown, HTML)
	err = r.enhanceRichReports(results, scenario, environment, metrics, cfg.Report.Thresholds)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// validateReportInputs checks that required inputs are provided and valid.
func (r *Reporter) validateReportInputs(resultFile, outputDir string) error {
	if resultFile == "" {
		return gerrors.ErrResultFileEmpty
	}

	if outputDir == "" {
		return gerrors.ErrOutputDirEmpty
	}

	// Check if result file exists
	if _, err := os.Stat(resultFile); os.IsNotExist(err) {
		return gerrors.WithResultFileNotFoundDetails(resultFile)
	}

	return nil
}

// getReportFormats returns the list of formats to generate.
func (r *Reporter) getReportFormats(cfg *config.Config) []string {
	formats := cfg.Report.Formats
	if len(formats) == 0 {
		// Default to text format if none specified
		formats = []string{FormatText}
	}
	return formats
}

// generateBasicReports creates basic reports for each requested format.
func (r *Reporter) generateBasicReports(
	resultFile, outputDir string,
	formats []string,
	metrics *Metrics,
	cfg *config.Config,
) ([]Result, error) {
	results := make([]Result, 0, len(formats))

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

		results = append(results, Result{
			FilePath: outputFile,
			Format:   format,
			Passed:   passed,
		})
	}

	return results, nil
}

// enhanceRichReports adds annotations to rich report formats like markdown and HTML.
func (r *Reporter) enhanceRichReports(
	results []Result,
	scenario, environment string,
	metrics *Metrics,
	thresholds map[string]string,
) error {
	for _, result := range results {
		if result.Format == FormatMarkdown {
			// Generate markdown with threshold annotations
			mdReport, err := r.GenerateMarkdownReport(scenario, environment, metrics, thresholds)
			if err != nil {
				return fmt.Errorf("failed to generate markdown report: %w", err)
			}

			// Write the markdown report
			if err := os.WriteFile(result.FilePath, []byte(mdReport), 0o600); err != nil {
				return fmt.Errorf("failed to write markdown report: %w", err)
			}
		} else if result.Format == FormatHTML {
			// Generate HTML with threshold annotations
			htmlReport, err := r.GenerateHTMLReport(scenario, environment, metrics, thresholds)
			if err != nil {
				return fmt.Errorf("failed to generate HTML report: %w", err)
			}

			// Write the HTML report
			if err := os.WriteFile(result.FilePath, []byte(htmlReport), 0o600); err != nil {
				return fmt.Errorf("failed to write HTML report: %w", err)
			}
		}
	}

	return nil
}

// GenerateReport generates a report in the specified format
func (r *Reporter) GenerateReport(resultFile, outputFile, format string, metrics *Metrics) error {
	var args []string
	var outputData []byte
	var err error

	// For markdown format, we handle it specially
	if format == FormatMarkdown && metrics != nil {
		// Generate basic markdown (will be enhanced later in GenerateReports)
		mdReport, err := r.GenerateMarkdownReport("", "", metrics, nil)
		if err != nil {
			return err
		}
		outputData = []byte(mdReport)
	} else if format == FormatHTML && metrics != nil {
		// Generate HTML report
		htmlReport, err := r.GenerateHTMLReport("", "", metrics, nil)
		if err != nil {
			return err
		}
		outputData = []byte(htmlReport)
	} else {
		// For other formats, use vegeta's report command
		args = []string{"report"}

		// Add format flag for non-text formats
		if format != FormatText {
			args = append(args, "-type="+format)
		}

		// Add input file
		args = append(args, resultFile)

		// Execute vegeta report command
		outputData, err = r.execCommand("vegeta", args...)
		if err != nil {
			return fmt.Errorf("vegeta report command failed: %w", err)
		}
	}

	// Write the report to file
	if err := os.WriteFile(outputFile, outputData, 0o600); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

// CheckThresholds validates metrics against configured thresholds.
func (r *Reporter) CheckThresholds(metrics *Metrics, thresholds map[string]string) (bool, error) {
	if metrics == nil {
		return true, nil
	}

	if len(thresholds) == 0 {
		return true, nil
	}

	violations := make([]string, 0)

	// Check different threshold types
	successRateViolations := r.checkSuccessRate(metrics, thresholds)
	violations = append(violations, successRateViolations...)

	latencyViolations := r.checkLatencyThresholds(metrics, thresholds)
	violations = append(violations, latencyViolations...)

	errorRateViolations := r.checkErrorRate(metrics, thresholds)
	violations = append(violations, errorRateViolations...)

	if len(violations) > 0 {
		return false, gerrors.WithThresholdViolationDetails(violations)
	}

	return true, nil
}

// checkSuccessRate validates the success rate against its threshold.
func (r *Reporter) checkSuccessRate(metrics *Metrics, thresholds map[string]string) []string {
	var violations []string

	if successRateThreshold, exists := thresholds["success_rate"]; exists {
		threshold, err := strconv.ParseFloat(successRateThreshold, 64)
		if err != nil {
			violations = append(violations,
				fmt.Sprintf("invalid success_rate threshold value: %s", successRateThreshold))
			return violations
		}

		if metrics.SuccessRate*100 < threshold {
			violations = append(violations,
				fmt.Sprintf("success_rate: %.2f%% < %.2f%%", metrics.SuccessRate*100, threshold))
		}
	}

	return violations
}

// checkLatencyThresholds validates various latency metrics against their thresholds.
func (r *Reporter) checkLatencyThresholds(metrics *Metrics, thresholds map[string]string) []string {
	var violations []string

	for metric, thresholdStr := range thresholds {
		var latencyNs int64
		var actualValue float64
		validMetric := true

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
			// Handled separately
			validMetric = false
		default:
			// Skip unknown metrics
			validMetric = false
		}

		if !validMetric {
			continue
		}

		// Parse threshold value (convert from "200ms" to milliseconds)
		thresholdValue, err := parseLatencyThreshold(thresholdStr)
		if err != nil {
			violations = append(violations,
				fmt.Sprintf("invalid threshold for %s: %s", metric, err))
			continue
		}

		// Compare actual value to threshold
		if actualValue > thresholdValue {
			violations = append(violations,
				fmt.Sprintf("%s: %.0fms > %.0fms", metric, actualValue, thresholdValue))
		}
	}

	return violations
}

// checkErrorRate validates the error rate against its threshold.
func (r *Reporter) checkErrorRate(metrics *Metrics, thresholds map[string]string) []string {
	var violations []string

	if errorRateThreshold, exists := thresholds["error_rate"]; exists {
		threshold, err := strconv.ParseFloat(errorRateThreshold, 64)
		if err != nil {
			violations = append(violations,
				fmt.Sprintf("invalid error_rate threshold value: %s", errorRateThreshold))
			return violations
		}

		errorRate := (1 - metrics.SuccessRate) * 100
		if errorRate > threshold {
			violations = append(violations,
				fmt.Sprintf("error_rate: %.2f%% > %.2f%%", errorRate, threshold))
		}
	}

	return violations
}

// GenerateMarkdownReport generates a markdown report with threshold annotations.
func (r *Reporter) GenerateMarkdownReport(
	scenario, environment string,
	metrics *Metrics,
	thresholds map[string]string,
) (string, error) {
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
	if len(thresholds) > 0 {
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
	output, err := r.execCommand("vegeta", "report", "-type=json", resultFile)
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

// ExtractMetrics extracts metrics from a Vegeta result file (public version)
func (r *Reporter) ExtractMetrics(resultFile string) (*Metrics, error) {
	return r.extractMetrics(resultFile)
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

// FormatExtension returns the file extension for a given format (public version)
func FormatExtension(format string) string {
	return formatExtension(format)
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
			return 0, gerrors.WithInvalidMsValueDetails(threshold)
		}
		return value, nil
	}

	if strings.HasSuffix(threshold, "s") {
		value, err := strconv.ParseFloat(strings.TrimSuffix(threshold, "s"), 64)
		if err != nil {
			return 0, gerrors.WithInvalidSecValueDetails(threshold)
		}
		return value * 1000, nil // Convert seconds to milliseconds
	}

	return 0, gerrors.WithUnknownLatencyUnitDetails(threshold)
}
