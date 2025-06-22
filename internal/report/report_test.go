package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kanywst/galick/internal/constants"
	"github.com/stretchr/testify/assert"
)

// TestGenerateReportFormats ensures that reports are generated correctly in various formats and that key metrics are reflected in the output. It also checks error handling for missing files.
// TestGenerateReportFormats tests that reports can be generated in different formats
func TestGenerateReportFormats(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Create a mock result file (this would normally be output from Vegeta)
	// This is a simplified binary version for testing
	resultFilePath := filepath.Join(tempDir, "results.bin")
	err = os.WriteFile(resultFilePath, []byte("mock-vegeta-data"), constants.FilePermissionDefault)
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
			if len(args) >= 4 && args[0] == "report" {
				inputFile := args[len(args)-1]
				if _, err := os.Stat(inputFile); os.IsNotExist(err) {
					return nil, err
				}
				outputFile := args[3]
				content := []byte(`{"success":1.0,"latencies":{"95th":150000000}}`)
				err := os.WriteFile(outputFile, content, constants.FilePermissionPrivate)
				if err != nil {
					return nil, err
				}
			}
			return []byte(`{"success":1.0,"latencies":{"95th":150000000}}`), nil
		},
		htmlTemplateCache: &HTMLTemplateCache{},
	}

	// Test JSON format (success case)
	jsonPath := filepath.Join(tempDir, "report.json")
	err = mockReporter.GenerateReport(resultFilePath, jsonPath, "json", nil)
	assert.NoError(t, err)
	jsonContent, err := os.ReadFile(jsonPath)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonContent), "success")

	// Test HTML format (success case)
	htmlPath := filepath.Join(tempDir, "report.html")
	err = mockReporter.GenerateReport(resultFilePath, htmlPath, "html", mockMetrics)
	assert.NoError(t, err)
	htmlContent, err := os.ReadFile(htmlPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, htmlContent)
	assert.Contains(t, string(htmlContent), "<html")
	assert.Contains(t, string(htmlContent), "99.5%") // Check if success rate is reflected in HTML
	assert.Contains(t, string(htmlContent), "120ms") // Check if P95 latency is reflected in HTML

	// Test Markdown format (success case)
	mdPath := filepath.Join(tempDir, "report.md")
	mdReport, err := mockReporter.GenerateMarkdownReport("Test Scenario", "Test Env", mockMetrics, nil)
	assert.NoError(t, err)
	err = os.WriteFile(mdPath, []byte(mdReport), constants.FilePermissionPrivate)
	assert.NoError(t, err)
	mdContent, err := os.ReadFile(mdPath)
	assert.NoError(t, err)
	assert.Contains(t, string(mdContent), "# Load Test Report")
	assert.Contains(t, string(mdContent), "99.5") // Check if success rate is reflected in Markdown
	assert.Contains(t, string(mdContent), "120") // Check if P95 latency is reflected in Markdown

	// Error case: Specify a non-existent file
	// Ensure that an error is returned when the input file does not exist
	err = mockReporter.GenerateReport("/notfound.bin", jsonPath, "json", nil)
	assert.Error(t, err)
}
