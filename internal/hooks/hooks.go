// Package hooks provides functionality for executing pre and post hooks
package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/kanywst/galick/internal/config"
)

// HookRunner executes pre and post hooks
type HookRunner struct {
	execCommand func(command string, args ...string) ([]byte, error)
}

// NewHookRunner creates a new hook runner
func NewHookRunner() *HookRunner {
	return &HookRunner{
		execCommand: func(command string, args ...string) ([]byte, error) {
			return exec.Command(command, args...).CombinedOutput()
		},
	}
}

// RunPreHook executes the pre-hook script if configured
func (h *HookRunner) RunPreHook(cfg *config.Config) error {
	if cfg.Hooks.Pre == "" {
		return nil
	}

	if _, err := os.Stat(cfg.Hooks.Pre); os.IsNotExist(err) {
		return fmt.Errorf("pre-hook script not found: %s", cfg.Hooks.Pre)
	}

	output, err := h.execCommand(cfg.Hooks.Pre)
	if err != nil {
		return fmt.Errorf("pre-hook script execution failed: %w\n%s", err, string(output))
	}

	return nil
}

// RunPostHook executes the post-hook script if configured
func (h *HookRunner) RunPostHook(cfg *config.Config, exitCode int) error {
	if cfg.Hooks.Post == "" {
		return nil
	}

	if _, err := os.Stat(cfg.Hooks.Post); os.IsNotExist(err) {
		return fmt.Errorf("post-hook script not found: %s", cfg.Hooks.Post)
	}

	output, err := h.execCommand(cfg.Hooks.Post, strconv.Itoa(exitCode))
	if err != nil {
		return fmt.Errorf("post-hook script execution failed: %w\n%s", err, string(output))
	}

	return nil
}
