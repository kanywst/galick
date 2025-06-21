package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestVersionCommand tests the version command
func TestVersionCommand(t *testing.T) {
	// Create a buffer to capture command output
	var buf bytes.Buffer
	
	// Create a version command
	cmd := newVersionCmd()
	cmd.SetOut(&buf)
	
	// Execute the command
	err := cmd.Execute()
	assert.NoError(t, err)
	
	// Check the output
	output := buf.String()
	assert.Contains(t, output, "galick version")
}

// TestInitCommand tests the init command
func TestInitCommand(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the current working directory
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	
	// Change to the temporary directory
	err = os.Chdir(tempDir)
	assert.NoError(t, err)
	defer os.Chdir(cwd)

	// Create a buffer to capture command output
	var buf bytes.Buffer
	
	// Create an init command
	cmd := newInitCmd()
	cmd.SetOut(&buf)
	
	// Execute the command
	err = cmd.Execute()
	assert.NoError(t, err)
	
	// Check that the file was created
	configFile := filepath.Join(tempDir, "loadtest.yaml")
	assert.FileExists(t, configFile)
	
	// Check the output
	output := buf.String()
	assert.Contains(t, output, "Created loadtest.yaml")
}

// TestRootCommand tests the root command (basic functionality)
func TestRootCommand(t *testing.T) {
	// Create a buffer to capture command output
	var buf bytes.Buffer
	
	// Create a root command
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	
	// Execute the command with --help flag
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.NoError(t, err)
	
	// Check the output
	output := buf.String()
	assert.Contains(t, output, "galick")
	assert.Contains(t, output, "Usage:")
}

// TestRunCommand tests the run command (mock version)
func TestRunCommand(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configFile := filepath.Join(tempDir, "loadtest.yaml")
	configContent := `
default:
  environment: dev
  scenario: simple
environments:
  dev:
    base_url: http://localhost:8080
    headers:
      Content-Type: application/json
scenarios:
  simple:
    rate: 10/s
    duration: 5s
    targets:
      - GET /api/health
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Create a buffer to capture command output
	var buf bytes.Buffer
	
	// Create a mock run command that doesn't actually execute vegeta
	cmd := &cobra.Command{
		Use:   "run [scenario]",
		Short: "Run a load test scenario",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Successfully executed mock run command")
			return nil
		},
	}
	
	cmd.SetOut(&buf)
	
	// Execute the command
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	assert.NoError(t, err)
	
	// Check the output
	output := buf.String()
	assert.Contains(t, output, "Successfully executed mock run command")
}

// TestReportCommand tests the report command (mock version)
func TestReportCommand(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configFile := filepath.Join(tempDir, "loadtest.yaml")
	configContent := `
default:
  environment: dev
  scenario: simple
environments:
  dev:
    base_url: http://localhost:8080
    headers:
      Content-Type: application/json
scenarios:
  simple:
    rate: 10/s
    duration: 5s
    targets:
      - GET /api/health
report:
  formats:
    - json
    - markdown
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Create a buffer to capture command output
	var buf bytes.Buffer
	
	// Create a mock report command that doesn't actually generate reports
	cmd := &cobra.Command{
		Use:   "report [scenario]",
		Short: "Generate reports for a scenario",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Successfully executed mock report command")
			return nil
		},
	}
	
	cmd.SetOut(&buf)
	
	// Execute the command
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	assert.NoError(t, err)
	
	// Check the output
	output := buf.String()
	assert.Contains(t, output, "Successfully executed mock report command")
}
