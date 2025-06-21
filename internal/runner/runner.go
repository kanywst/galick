// Package runner provides functionality to execute load tests using Vegeta.
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

// Runner executes load test scenarios.
type Runner struct {
	execCommand func(cmd string, args ...string) ([]byte, error)
}

// Result contains the result of a load test run.
type Result struct {
	OutputFile   string
	Environment  string
	Scenario     string
	OutputFolder string
}

// NewRunner creates a new runner instance.
func NewRunner() *Runner {
	return &Runner{
		execCommand: func(cmd string, args ...string) ([]byte, error) {
			return exec.Command(cmd, args...).CombinedOutput()
		},
	}
}

// RunScenario executes a load test scenario.
func (r *Runner) RunScenario(cfg *config.Config, scenarioName, environmentName, outputFolder string) (*Result, error) {
	// Get the scenario and environment.
	scenario, err := cfg.GetScenario(scenarioName)
	if err != nil {
		return nil, err
	}

	environment, err := cfg.GetEnvironment(environmentName)
	if err != nil {
		return nil, err
	}

	// Create the output directory
	if err := os.MkdirAll(outputFolder, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Output file path.
	outputFile := filepath.Join(outputFolder, "results.bin")

	// Build and execute the Vegeta command.
	cmd, targetsFile, err := r.buildVegetaCommand(scenario, environment, outputFile)
	if err != nil {
		return nil, err
	}
	defer os.Remove(targetsFile)

	// Execute the command.
	output, err := r.execCommand(cmd.Path, cmd.Args[1:]...)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("vegeta attack failed with exit code %d: %s", exitError.ExitCode(), string(output))
		}
		// Check if the command wasn't found.
		if strings.Contains(err.Error(), "executable file not found") {
			return nil, fmt.Errorf(
				"vegeta command not found. Please install Vegeta " +
					"(https://github.com/tsenart/vegeta) and make sure it's in your PATH",
			)
		}
		// Check if output contains permission denied.
		if strings.Contains(string(output), "permission denied") {
			return nil, fmt.Errorf("permission denied when executing vegeta. Make sure vegeta is executable")
		}
		return nil, fmt.Errorf("vegeta attack failed: %w\n%s", err, string(output))
	}

	// Check if output file was created.
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("output file was not created, vegeta may have failed silently")
	}

	return &Result{
		OutputFile:   outputFile,
		Environment:  environmentName,
		Scenario:     scenarioName,
		OutputFolder: outputFolder,
	}, nil
}

// buildVegetaCommand constructs the Vegeta command and creates a temporary targets file.
func (r *Runner) buildVegetaCommand(
	scenario *config.Scenario,
	environment *config.Environment,
	outputFile string,
) (*exec.Cmd, string, error) {
	// Create a temporary file for targets.
	targetsFile, err := r.createTargetsFile(scenario.Targets, environment)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create targets file: %w", err)
	}

	// Build the Vegeta command.
	args := []string{
		"attack",
		"-targets=" + targetsFile,
		"-rate=" + scenario.Rate,
		"-duration=" + scenario.Duration,
		"-output=" + outputFile,
	}

	// Use a fixed binary path for security.
	cmd := exec.Command("vegeta", args...)

	return cmd, targetsFile, nil
}

// createTargetsFile creates a temporary file with the target URLs and headers.
func (r *Runner) createTargetsFile(targets []string, environment *config.Environment) (string, error) {
	// Create a temporary file
	file, err := os.CreateTemp("", "vegeta-targets-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary targets file: %w", err)
	}
	defer file.Close()

	// Validate environment.
	if environment.BaseURL == "" {
		return "", fmt.Errorf("environment base URL is empty")
	}

	// Write targets to the file.
	for i, target := range targets {
		parts := strings.SplitN(target, " ", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid target format at index %d: %s (expected 'METHOD /path')", i, target)
		}

		method := parts[0]
		path := parts[1]

		// Validate method.
		method = strings.ToUpper(method)
		validMethods := map[string]bool{
			"GET":     true,
			"POST":    true,
			"PUT":     true,
			"DELETE":  true,
			"PATCH":   true,
			"HEAD":    true,
			"OPTIONS": true,
		}
		if !validMethods[method] {
			return "", fmt.Errorf("invalid HTTP method at index %d: %s", i, method)
		}

		// Ensure the path starts with "/".
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		// Build the full URL.
		url := environment.BaseURL
		if strings.HasSuffix(url, "/") && strings.HasPrefix(path, "/") {
			// Avoid double slash if base URL ends with / and path starts with "/".
			url = url + strings.TrimPrefix(path, "/")
		} else if !strings.HasSuffix(url, "/") && !strings.HasPrefix(path, "/") {
			// Add slash if neither has it.
			url = url + "/" + path
		} else {
			// Either base ends with / or path starts with / but not both.
			url = url + path
		}

		// Write the target line.
		_, err := fmt.Fprintf(file, "%s %s\n", method, url)
		if err != nil {
			return "", fmt.Errorf("failed to write target to file: %w", err)
		}

		// Write headers.
		for key, value := range environment.Headers {
			_, err := fmt.Fprintf(file, "%s: %s\n", key, value)
			if err != nil {
				return "", fmt.Errorf("failed to write header to file: %w", err)
			}
		}
		_, err = fmt.Fprintln(file) // Empty line to separate targets.
		if err != nil {
			return "", fmt.Errorf("failed to write line separator to file: %w", err)
		}
	}

	return file.Name(), nil
}

// RunPreHook executes the pre-hook script if configured.
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

// RunPostHook executes the post-hook script if configured.
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
