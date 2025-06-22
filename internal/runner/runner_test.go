package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kanywst/galick/internal/config"
	"github.com/kanywst/galick/internal/constants"
	"github.com/stretchr/testify/assert"
)

// TestBuildVegetaCommand tests that the Vegeta command is correctly built.
func TestBuildVegetaCommand(t *testing.T) {
	// Create a test scenario and environment
	scenario := &config.Scenario{
		Rate:     "50/s",
		Duration: "30s",
		Targets: []string{
			"GET /api/users",
			"POST /api/products",
		},
	}

	environment := &config.Environment{
		BaseURL: "http://example.com",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
	}

	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir) // エラーは無視
	}()

	outputPath := filepath.Join(tempDir, "output.bin")

	// Build the Vegeta command
	runner := NewRunner()
	cmd, targetsFile, err := runner.buildVegetaCommand(scenario, environment, outputPath)
	assert.NoError(t, err)
	defer func() {
		_ = os.Remove(targetsFile) // エラーは無視
	}()

	// Verify command arguments
	assert.Contains(t, cmd.Args, "attack")
	assert.Contains(t, cmd.Args, "-rate=50/s")
	assert.Contains(t, cmd.Args, "-duration=30s")
	assert.Contains(t, cmd.Args, "-output="+outputPath)

	// Verify targets file content
	// ファイルパスの検証（安全性確認）
	if !filepath.IsAbs(targetsFile) {
		t.Fatalf("Expected absolute path, got: %s", targetsFile)
	}
	content, err := os.ReadFile(targetsFile)
	assert.NoError(t, err)

	// The targets file should contain the fully-formed URLs with headers
	targetContent := string(content)
	assert.Contains(t, targetContent, "GET http://example.com/api/users")
	assert.Contains(t, targetContent, "POST http://example.com/api/products")
	assert.Contains(t, targetContent, "Content-Type: application/json")
	assert.Contains(t, targetContent, "Authorization: Bearer token123")
}

// TestRunScenario tests the scenario execution (mock version).
func TestRunScenario(t *testing.T) {
	// Create a mock config
	cfg := &config.Config{
		Default: config.Default{
			Environment: "dev",
			Scenario:    "simple",
			OutputDir:   "output",
		},
		Environments: map[string]config.Environment{
			"dev": {
				BaseURL: "http://localhost:8080",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		Scenarios: map[string]config.Scenario{
			"simple": {
				Rate:     "10/s",
				Duration: "5s",
				Targets:  []string{"GET /api/health"},
			},
		},
	}

	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir) // エラーは無視
	}()

	// Create a mock runner that doesn't actually execute commands
	mockRunner := &Runner{
		execCommand: func(_ string, _ ...string) ([]byte, error) {
			// Create a dummy results file to simulate successful execution
			resultsFile := filepath.Join(tempDir, "results.bin")
			err := os.WriteFile(resultsFile, []byte("mock vegeta binary data"), constants.FilePermissionDefault)
			if err != nil {
				return nil, err
			}
			return []byte("Mocked execution"), nil
		},
	}

	// Run the scenario
	result, err := mockRunner.RunScenario(cfg, "simple", "dev", tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Join(tempDir, "results.bin"), result.OutputFile)
}

// TestRunPreHook tests the pre-hook execution.
func TestRunPreHook(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir) // エラーは無視
	}()

	// Create a test script
	scriptContent := `#!/bin/sh
echo "Pre-hook executed"
exit 0
`
	scriptPath := filepath.Join(tempDir, "pre-hook.sh")
	err = os.WriteFile(scriptPath, []byte(scriptContent), constants.FilePermissionPrivate)
	assert.NoError(t, err)

	// Create a mock config with hooks
	cfg := &config.Config{
		Hooks: config.Hooks{
			Pre: scriptPath,
		},
	}

	// Create a runner
	runner := NewRunner()

	// Run the pre-hook
	err = runner.RunPreHook(cfg)
	assert.NoError(t, err)

	// Test with a non-existent script
	cfg.Hooks.Pre = filepath.Join(tempDir, "nonexistent.sh")
	err = runner.RunPreHook(cfg)
	assert.Error(t, err)
}

// TestRunPostHook tests the post-hook execution.
func TestRunPostHook(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir) // エラーは無視
	}()

	// Create a test script
	scriptContent := `#!/bin/sh
echo "Post-hook executed with exit code: $1"
exit 0
`
	scriptPath := filepath.Join(tempDir, "post-hook.sh")
	err = os.WriteFile(scriptPath, []byte(scriptContent), constants.FilePermissionPrivate)
	assert.NoError(t, err)

	// Create a mock config with hooks
	cfg := &config.Config{
		Hooks: config.Hooks{
			Post: scriptPath,
		},
	}

	// Create a runner
	runner := NewRunner()

	// Run the post-hook with exit code 0
	err = runner.RunPostHook(cfg, 0)
	assert.NoError(t, err)

	// Run the post-hook with exit code 1
	err = runner.RunPostHook(cfg, 1)
	assert.NoError(t, err)

	// Test with a non-existent script
	cfg.Hooks.Post = filepath.Join(tempDir, "nonexistent.sh")
	err = runner.RunPostHook(cfg, 0)
	assert.Error(t, err)
}
