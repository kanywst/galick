// Package cli_test provides tests for the cli package
package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kanywst/galick/internal/cli"
	"github.com/stretchr/testify/assert"
)

// TestVersionCommand tests the version command.
func TestVersionCommand(t *testing.T) {
	// Create a buffer to capture command output
	var buf bytes.Buffer

	// Create a root command with version subcommand
	app := cli.NewApp()
	cmd := app.NewRootCmd()
	cmd.SetOut(&buf)

	// Set args to simulate version command
	cmd.SetArgs([]string{"version"})

	// Execute the command
	err := cmd.Execute()
	assert.NoError(t, err)

	// Check the output
	output := buf.String()
	assert.Contains(t, output, "galick version")
}

// TestInitCommand tests the init command.
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
	defer func() {
		err := os.Chdir(cwd)
		if err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()

	// Create a buffer to capture command output
	var buf bytes.Buffer

	// Create a root command with init subcommand
	app := cli.NewApp()
	cmd := app.NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Mock stdin to simulate user input for overwrite prompt (if needed)
	// This is not needed for first run, but in case the file exists
	// in the future, we'll have a response ready
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, err = w.Write([]byte("y\n"))
	if err != nil {
		t.Errorf("Failed to write to pipe: %v", err)
	}
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	// Set args to simulate init command
	cmd.SetArgs([]string{"init"})

	// Execute the command
	err = cmd.Execute()
	assert.NoError(t, err)

	// Check if the config file was created
	configFile := filepath.Join(tempDir, "loadtest.yaml")
	_, err = os.Stat(configFile)
	assert.NoError(t, err)

	// Check the output
	output := buf.String()
	assert.Contains(t, output, "Created config file")
}

// TestRootCommand tests the root command.
func TestRootCommand(t *testing.T) {
	// Create a buffer to capture command output
	var buf bytes.Buffer

	// Create a root command
	app := cli.NewApp()
	cmd := app.NewRootCmd()
	cmd.SetOut(&buf)

	// Execute the command with help flag
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Check the output
	output := buf.String()
	assert.Contains(t, output, "galick")
	assert.Contains(t, output, "Available Commands")
}

// TestRunCommand tests the run command (minimal test).
func TestRunCommand(t *testing.T) {
	// This is a minimal test that just checks command creation
	// A full test would require mocking the runner and config

	// Create a root command with run subcommand
	app := cli.NewApp()
	cmd := app.NewRootCmd()
	runCmd, _, err := cmd.Find([]string{"run"})

	// Check that the command exists and has the expected properties
	assert.NoError(t, err)
	assert.NotNil(t, runCmd)
	assert.Equal(t, "run", runCmd.Name())
	assert.Contains(t, runCmd.Short, "Run a load test scenario")
}

// TestReportCommand tests the report command (minimal test)
func TestReportCommand(t *testing.T) {
	// This is a minimal test that just checks command creation
	// A full test would require mocking the reporter and config

	// Create a root command with report subcommand
	app := cli.NewApp()
	cmd := app.NewRootCmd()
	reportCmd, _, err := cmd.Find([]string{"report"})

	// Check that the command exists and has the expected properties
	assert.NoError(t, err)
	assert.NotNil(t, reportCmd)
	assert.Equal(t, "report", reportCmd.Name())
	assert.Contains(t, reportCmd.Short, "Generate a report")
}
