package report

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	gerrors "github.com/kanywst/galick/internal/errors"
)

//go:embed templates/html_report.tmpl
var htmlReportTemplate string

// HTMLData contains data for HTML report template.
type HTMLData struct {
	Environment          string
	Scenario             string
	Date                 string
	Metrics              *Metrics
	Duration             string
	Passed               bool
	SuccessRate          string
	SuccessRateStatus    bool
	SuccessRateThreshold string
	P95Latency           string
	P95Status            bool
	P95Threshold         string
	Throughput           string
	MeanLatency          string
	ThresholdResults     []ThresholdResult
}

// ThresholdResult represents a single threshold check result.
type ThresholdResult struct {
	Metric    string
	Threshold string
	Actual    string
	Passed    bool
}

// GenerateHTMLReport creates an HTML report from the metrics and thresholds.
func (r *Reporter) GenerateHTMLReport(
	scenario, environment string,
	metrics *Metrics,
	thresholds map[string]string,
) (string, error) {
	if metrics == nil {
		return "", gerrors.ErrMetricsNil
	}

	// Parse the template
	tmpl, err := template.New("html").Parse(htmlReportTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Check thresholds and prepare data
	passed := true
	thresholdResults := []ThresholdResult{}

	// Success Rate
	successRateThreshold, successRateStatus, successRateResult := r.checkSuccessRateThreshold(metrics, thresholds)
	thresholdResults = append(thresholdResults, successRateResult)
	if !successRateStatus {
		passed = false
	}

	// P95 Latency
	p95Threshold, p95Status, p95Result := r.checkP95LatencyThreshold(metrics, thresholds)
	thresholdResults = append(thresholdResults, p95Result)
	if !p95Status {
		passed = false
	}

	// Other latency thresholds
	otherThresholdResults, allPassed := r.checkOtherLatencyThresholds(metrics, thresholds)
	thresholdResults = append(thresholdResults, otherThresholdResults...)
	if !allPassed {
		passed = false
	}

	// Prepare template data
	data := HTMLData{
		Environment:          environment,
		Scenario:             scenario,
		Date:                 time.Now().Format("2006-01-02 15:04:05"),
		Metrics:              metrics,
		Duration:             fmt.Sprintf("%.1fs", float64(metrics.Duration)/1000000000),
		Passed:               passed,
		SuccessRate:          fmt.Sprintf("%.1f", metrics.SuccessRate*100),
		SuccessRateStatus:    successRateStatus,
		SuccessRateThreshold: successRateThreshold,
		P95Latency:           fmt.Sprintf("%.0fms", float64(metrics.Latencies.P95)/1000000),
		P95Status:            p95Status,
		P95Threshold:         p95Threshold,
		Throughput:           fmt.Sprintf("%.1f", metrics.Throughput),
		MeanLatency:          fmt.Sprintf("%.0fms", float64(metrics.Latencies.Mean)/1000000),
		ThresholdResults:     thresholdResults,
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return buf.String(), nil
}

// checkSuccessRateThreshold checks the success rate against its threshold.
func (r *Reporter) checkSuccessRateThreshold(
	metrics *Metrics,
	thresholds map[string]string,
) (string, bool, ThresholdResult) {
	successRateThreshold := ""
	successRateStatus := true
	var thresholdResult ThresholdResult

	if val, ok := thresholds["success_rate"]; ok {
		threshold, err := parseThresholdValue(val)
		if err == nil {
			successRateThreshold = val
			actual := metrics.SuccessRate * 100
			if actual < threshold {
				successRateStatus = false
			}
			thresholdResult = ThresholdResult{
				Metric:    "Success Rate",
				Threshold: fmt.Sprintf("%.1f%%", threshold),
				Actual:    fmt.Sprintf("%.1f%%", actual),
				Passed:    successRateStatus,
			}
		}
	}

	return successRateThreshold, successRateStatus, thresholdResult
}

// checkP95LatencyThreshold checks the P95 latency against its threshold.
func (r *Reporter) checkP95LatencyThreshold(
	metrics *Metrics,
	thresholds map[string]string,
) (string, bool, ThresholdResult) {
	p95Threshold := ""
	p95Status := true
	var thresholdResult ThresholdResult

	if val, ok := thresholds["p95"]; ok {
		threshold, err := parseLatencyThreshold(val)
		if err == nil {
			p95Threshold = val
			actual := float64(metrics.Latencies.P95) / 1000000 // ns to ms
			if actual > threshold {
				p95Status = false
			}
			thresholdResult = ThresholdResult{
				Metric:    "P95 Latency",
				Threshold: val,
				Actual:    fmt.Sprintf("%.0fms", actual),
				Passed:    p95Status,
			}
		}
	}

	return p95Threshold, p95Status, thresholdResult
}

// checkOtherLatencyThresholds checks other latency metrics against their thresholds.
func (r *Reporter) checkOtherLatencyThresholds(
	metrics *Metrics,
	thresholds map[string]string,
) ([]ThresholdResult, bool) {
	thresholdResults := []ThresholdResult{}
	allPassed := true

	latencyMetrics := map[string]int64{
		"p50":  metrics.Latencies.P50,
		"p90":  metrics.Latencies.P90,
		"p99":  metrics.Latencies.P99,
		"mean": metrics.Latencies.Mean,
		"max":  metrics.Latencies.Max,
	}

	for metric, value := range latencyMetrics {
		if val, ok := thresholds[metric]; ok && metric != "p95" { // p95 already handled
			threshold, err := parseLatencyThreshold(val)
			if err == nil {
				actual := float64(value) / 1000000 // ns to ms
				metricPassed := actual <= threshold
				if !metricPassed {
					allPassed = false
				}
				thresholdResults = append(thresholdResults, ThresholdResult{
					Metric:    strings.ToUpper(metric) + " Latency",
					Threshold: val,
					Actual:    fmt.Sprintf("%.0fms", actual),
					Passed:    metricPassed,
				})
			}
		}
	}

	return thresholdResults, allPassed
}

// parseThresholdValue parses a threshold value string.
func parseThresholdValue(value string) (float64, error) {
	// Try to parse as a float
	return parseLatencyThreshold(value)
}

// SaveHTMLReport generates and saves an HTML report for a test result.
func (r *Reporter) SaveHTMLReport(
	resultFile, outputPath, scenario, environment string,
	thresholds map[string]string,
) (string, error) {
	// Extract metrics from the result file
	metrics, err := r.extractMetrics(resultFile)
	if err != nil {
		return "", fmt.Errorf("failed to extract metrics: %w", err)
	}

	// Generate the HTML report
	html, err := r.GenerateHTMLReport(scenario, environment, metrics, thresholds)
	if err != nil {
		return "", fmt.Errorf("failed to generate HTML report: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the HTML to file
	if err := os.WriteFile(outputPath, []byte(html), 0o600); err != nil {
		return "", fmt.Errorf("failed to write HTML report: %w", err)
	}

	return outputPath, nil
}
