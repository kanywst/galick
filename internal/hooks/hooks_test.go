package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kanywst/galick/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestRunPreHook tests the execution of pre-hooks
func TestRunPreHook(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test script
	scriptContent := `#!/bin/sh
echo "Pre-hook executed"
exit 0
`
	scriptPath := filepath.Join(tempDir, "pre-hook.sh")
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	assert.NoError(t, err)

	// Create a mock config with hooks
	cfg := &config.Config{
		Hooks: config.Hooks{
			Pre: scriptPath,
		},
	}

	// Create a hook runner with mock execution
	var executedCommand string
	var executedArgs []string

	runner := &HookRunner{
		execCommand: func(command string, args ...string) ([]byte, error) {
			executedCommand = command
			executedArgs = args
			return []byte("Executed successfully"), nil
		},
	}

	// Run the pre-hook
	err = runner.RunPreHook(cfg)
	assert.NoError(t, err)
	assert.Equal(t, scriptPath, executedCommand)
	assert.Empty(t, executedArgs)

	// Test with a non-existent hook
	cfg.Hooks.Pre = filepath.Join(tempDir, "nonexistent.sh")
	err = runner.RunPreHook(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestRunPostHook tests the execution of post-hooks
func TestRunPostHook(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "galick-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test script
	scriptContent := `#!/bin/sh
echo "Post-hook executed with exit code: $1"
exit 0
`
	scriptPath := filepath.Join(tempDir, "post-hook.sh")
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	assert.NoError(t, err)

	// Create a mock config with hooks
	cfg := &config.Config{
		Hooks: config.Hooks{
			Post: scriptPath,
		},
	}

	// Create a hook runner with mock execution
	var executedCommand string
	var executedArgs []string

	runner := &HookRunner{
		execCommand: func(command string, args ...string) ([]byte, error) {
			executedCommand = command
			executedArgs = args
			return []byte("Executed successfully"), nil
		},
	}

	// Run the post-hook with exit code 0
	err = runner.RunPostHook(cfg, 0)
	assert.NoError(t, err)
	assert.Equal(t, scriptPath, executedCommand)
	assert.Equal(t, []string{"0"}, executedArgs)

	// Run the post-hook with exit code 1
	err = runner.RunPostHook(cfg, 1)
	assert.NoError(t, err)
	assert.Equal(t, scriptPath, executedCommand)
	assert.Equal(t, []string{"1"}, executedArgs)

	// Test with a non-existent hook
	cfg.Hooks.Post = filepath.Join(tempDir, "nonexistent.sh")
	err = runner.RunPostHook(cfg, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
