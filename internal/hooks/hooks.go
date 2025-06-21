// Package hooks provides functionality for executing pre and post hooks.
package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/kanywst/galick/internal/config"
)

// HookRunner executes pre and post hooks.
type HookRunner struct {
	execCommand func(command string, args ...string) ([]byte, error)
}

// NewHookRunner creates a new hook runner.
func NewHookRunner() *HookRunner {
	return &HookRunner{
		execCommand: func(command string, args ...string) ([]byte, error) {
			return exec.Command(command, args...).CombinedOutput()
		},
	}
}

// RunPreHook executes the pre-hook script if configured.
func (h *HookRunner) RunPreHook(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if cfg.Hooks.Pre == "" {
		// No pre-hook defined, this is normal and not an error
		return nil
	}

	// Check if the script exists and is executable
	info, err := os.Stat(cfg.Hooks.Pre)
	if os.IsNotExist(err) {
		return fmt.Errorf("pre-hook script not found: %s", cfg.Hooks.Pre)
	}

	if err != nil {
		return fmt.Errorf("failed to check pre-hook script: %w", err)
	}

	// On Unix systems, check if the script is executable
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("pre-hook script is not executable: %s", cfg.Hooks.Pre)
	}

	output, err := h.execCommand(cfg.Hooks.Pre)
	if err != nil {
		return fmt.Errorf("pre-hook script execution failed: %w\n%s", err, string(output))
	}

	return nil
}

// RunPostHook executes the post-hook script if configured.
func (h *HookRunner) RunPostHook(cfg *config.Config, exitCode int) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if cfg.Hooks.Post == "" {
		// No post-hook defined, this is normal and not an error
		return nil
	}

	// Check if the script exists and is executable
	info, err := os.Stat(cfg.Hooks.Post)
	if os.IsNotExist(err) {
		return fmt.Errorf("post-hook script not found: %s", cfg.Hooks.Post)
	}

	if err != nil {
		return fmt.Errorf("failed to check post-hook script: %w", err)
	}

	// On Unix systems, check if the script is executable
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("post-hook script is not executable: %s", cfg.Hooks.Post)
	}

	output, err := h.execCommand(cfg.Hooks.Post, strconv.Itoa(exitCode))
	if err != nil {
		return fmt.Errorf("post-hook script execution failed: %w\n%s", err, string(output))
	}

	return nil
}
