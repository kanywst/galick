// Package runner provides functionality to execute load tests using Vegeta
package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kanywst/galick/internal/config"
)

// Runner executes load test scenarios
type Runner struct {
	execCommand func(cmd string, args ...string) ([]byte, error)
}

// Result contains the result of a load test run
type Result struct {
	OutputFile   string
	Environment  string
	Scenario     string
	OutputFolder string
}

// NewRunner creates a new runner instance
func NewRunner() *Runner {
	return &Runner{
		execCommand: func(cmd string, args ...string) ([]byte, error) {
			return exec.Command(cmd, args...).CombinedOutput()
		},
	}
}

// RunScenario executes a load test scenario
func (r *Runner) RunScenario(cfg *config.Config, scenarioName, environmentName, outputFolder string) (*Result, error) {
	// Get the scenario and environment
	scenario, err := cfg.GetScenario(scenarioName)
	if err != nil {
		return nil, err
	}

	environment, err := cfg.GetEnvironment(environmentName)
	if err != nil {
		return nil, err
	}

	// Create the output directory
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Output file path
	outputFile := filepath.Join(outputFolder, "results.bin")

	// Build and execute the Vegeta command
	cmd, targetsFile, err := r.buildVegetaCommand(scenario, environment, outputFile)
	if err != nil {
		return nil, err
	}
	defer os.Remove(targetsFile)

	// Execute the command
	output, err := r.execCommand(cmd.Path, cmd.Args[1:]...)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("vegeta attack failed with exit code %d: %s", exitError.ExitCode(), string(output))
		}
		// Check if the command wasn't found
		if strings.Contains(err.Error(), "executable file not found") {
			return nil, fmt.Errorf("vegeta command not found. Please install Vegeta (https://github.com/tsenart/vegeta) and make sure it's in your PATH")
		}
		return nil, fmt.Errorf("vegeta attack failed: %w\n%s", err, string(output))
	}

	return &Result{
		OutputFile:   outputFile,
		Environment:  environmentName,
		Scenario:     scenarioName,
		OutputFolder: outputFolder,
	}, nil
}

// buildVegetaCommand constructs the Vegeta command and creates a temporary targets file
func (r *Runner) buildVegetaCommand(scenario *config.Scenario, environment *config.Environment, outputFile string) (*exec.Cmd, string, error) {
	// Create a temporary file for targets
	targetsFile, err := r.createTargetsFile(scenario.Targets, environment)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create targets file: %w", err)
	}

	// Build the Vegeta command
	cmd := exec.Command(
		"vegeta",
		"attack",
		"-targets="+targetsFile,
		"-rate="+scenario.Rate,
		"-duration="+scenario.Duration,
		"-output="+outputFile,
	)

	return cmd, targetsFile, nil
}

// createTargetsFile creates a temporary file with the target URLs and headers
func (r *Runner) createTargetsFile(targets []string, environment *config.Environment) (string, error) {
	// Create a temporary file
	file, err := os.CreateTemp("", "vegeta-targets-*.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Write targets to the file
	for _, target := range targets {
		parts := strings.SplitN(target, " ", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid target format: %s (expected 'METHOD /path')", target)
		}

		method := parts[0]
		path := parts[1]

		// Ensure the path starts with /
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		// Build the full URL
		url := environment.BaseURL
		if strings.HasSuffix(url, "/") && strings.HasPrefix(path, "/") {
			// Avoid double slash if base URL ends with / and path starts with /
			url = url + strings.TrimPrefix(path, "/")
		} else if !strings.HasSuffix(url, "/") && !strings.HasPrefix(path, "/") {
			// Add slash if neither has it
			url = url + "/" + path
		} else {
			// Either base ends with / or path starts with / but not both
			url = url + path
		}

		// Write the target line
		fmt.Fprintf(file, "%s %s\n", method, url)

		// Write headers
		for key, value := range environment.Headers {
			fmt.Fprintf(file, "%s: %s\n", key, value)
		}
		fmt.Fprintln(file) // Empty line to separate targets
	}

	return file.Name(), nil
}

// RunPreHook executes the pre-hook script if configured
func (r *Runner) RunPreHook(cfg *config.Config) error {
	if cfg.Hooks.Pre == "" {
		return nil
	}

	if _, err := os.Stat(cfg.Hooks.Pre); os.IsNotExist(err) {
		return fmt.Errorf("pre-hook script not found: %s", cfg.Hooks.Pre)
	}

	output, err := r.execCommand(cfg.Hooks.Pre)
	if err != nil {
		return fmt.Errorf("pre-hook script execution failed: %w\n%s", err, string(output))
	}

	return nil
}

// RunPostHook executes the post-hook script if configured
func (r *Runner) RunPostHook(cfg *config.Config, exitCode int) error {
	if cfg.Hooks.Post == "" {
		return nil
	}

	if _, err := os.Stat(cfg.Hooks.Post); os.IsNotExist(err) {
		return fmt.Errorf("post-hook script not found: %s", cfg.Hooks.Post)
	}

	output, err := r.execCommand(cfg.Hooks.Post, strconv.Itoa(exitCode))
	if err != nil {
		return fmt.Errorf("post-hook script execution failed: %w\n%s", err, string(output))
	}

	return nil
}
