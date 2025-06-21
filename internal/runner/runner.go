// Package runner provides functionality to execute load tests using Vegeta.
package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kanywst/galick/internal/config"
	"github.com/kanywst/galick/internal/constants"
	gerrors "github.com/kanywst/galick/internal/errors"
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
	if err := os.MkdirAll(outputFolder, constants.DirPermissionDefault); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Output file path.
	outputFile := filepath.Join(outputFolder, "results.bin")

	// Prepare and execute the load test
	if err := r.executeLoadTest(scenario, environment, outputFile); err != nil {
		return nil, err
	}

	return &Result{
		OutputFile:   outputFile,
		Environment:  environmentName,
		Scenario:     scenarioName,
		OutputFolder: outputFolder,
	}, nil
}

// executeLoadTest prepares and executes the Vegeta command.
func (r *Runner) executeLoadTest(scenario *config.Scenario, environment *config.Environment, outputFile string) error {
	// Build and execute the Vegeta command.
	cmd, targetsFile, err := r.buildVegetaCommand(scenario, environment, outputFile)
	if err != nil {
		return err
	}
	defer os.Remove(targetsFile)

	// Execute the command.
	output, err := r.execCommand(cmd.Path, cmd.Args[1:]...)
	if err != nil {
		return r.handleVegetaError(err, output)
	}

	// Check if output file was created.
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return gerrors.ErrOutputNotCreated
	}

	return nil
}

// handleVegetaError processes different types of errors from Vegeta execution.
func (r *Runner) handleVegetaError(err error, output []byte) error {
	switch {
	case r.isExitError(err):
		var exitErr *exec.ExitError
		errors.As(err, &exitErr)
		return fmt.Errorf("%w with exit code %d: %s", gerrors.ErrVegetaAttackFailed, exitErr.ExitCode(), string(output))
	case r.isCommandNotFound(err):
		return gerrors.ErrVegetaNotFound
	case r.isPermissionDenied(string(output)):
		return gerrors.ErrVegetaNotExec
	default:
		return fmt.Errorf("%w: %w\n%s", gerrors.ErrVegetaAttackFailed, err, string(output))
	}
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
		return "", gerrors.ErrEnvironmentBaseURLEmpty
	}

	// Write targets to the file.
	for i, target := range targets {
		method, path, err := r.validateTarget(target, i)
		if err != nil {
			return "", err
		}

		url := r.buildFullURL(environment.BaseURL, path)

		err = r.writeTargetToFile(file, method, url, environment.Headers)
		if err != nil {
			return "", err
		}
	}

	return file.Name(), nil
}

// validateTarget validates a target string format.
func (r *Runner) validateTarget(target string, index int) (string, string, error) {
	parts := strings.SplitN(target, " ", constants.DefaultSplitParts)
	if len(parts) != constants.DefaultSplitParts {
		return "", "", gerrors.WithInvalidTargetFormatDetails(index, target)
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
		return "", "", gerrors.WithInvalidHTTPMethodDetails(index, method)
	}

	// Ensure the path starts with "/".
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return method, path, nil
}

// buildFullURL builds a full URL from a base URL and path.
func (r *Runner) buildFullURL(baseURL, path string) string {
	url := baseURL

	switch {
	case strings.HasSuffix(url, "/") && strings.HasPrefix(path, "/"):
		// Avoid double slash if base URL ends with / and path starts with "/".
		url += strings.TrimPrefix(path, "/")
	case !strings.HasSuffix(url, "/") && !strings.HasPrefix(path, "/"):
		// Add slash if neither has it.
		url += "/" + path
	default:
		// Either base ends with / or path starts with / but not both.
		url += path
	}
	return url
}

// writeTargetToFile writes a single target with its headers to the targets file.
func (r *Runner) writeTargetToFile(file *os.File, method, url string, headers map[string]string) error {
	// Write the target line.
	if err := r.writeLineToFile(file, fmt.Sprintf("%s %s", method, url)); err != nil {
		return err
	}

	// Write headers.
	for key, value := range headers {
		if err := r.writeLineToFile(file, fmt.Sprintf("%s: %s", key, value)); err != nil {
			return err
		}
	}

	// Empty line to separate targets.
	if err := r.writeLineToFile(file, ""); err != nil {
		return err
	}

	return nil
}

// writeLineToFile writes a single line to a file with error handling.
func (r *Runner) writeLineToFile(file *os.File, line string) error {
	var err error
	if line == "" {
		_, err = fmt.Fprintln(file)
	} else {
		_, err = fmt.Fprintln(file, line)
	}

	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

// RunPreHook executes the pre-hook script if configured.
func (r *Runner) RunPreHook(cfg *config.Config) error {
	if cfg.Hooks.Pre == "" {
		return nil
	}

	if _, err := os.Stat(cfg.Hooks.Pre); os.IsNotExist(err) {
		return gerrors.WithPreHookNotFoundDetails(cfg.Hooks.Pre)
	}

	output, err := r.execCommand(cfg.Hooks.Pre)
	if err != nil {
		return fmt.Errorf("%w: %w\n%s", gerrors.ErrPreHookNotExec, err, string(output))
	}

	return nil
}

// RunPostHook executes the post-hook script if configured.
func (r *Runner) RunPostHook(cfg *config.Config, exitCode int) error {
	if cfg.Hooks.Post == "" {
		return nil
	}

	if _, err := os.Stat(cfg.Hooks.Post); os.IsNotExist(err) {
		return gerrors.WithPostHookNotFoundDetails(cfg.Hooks.Post)
	}

	output, err := r.execCommand(cfg.Hooks.Post, strconv.Itoa(exitCode))
	if err != nil {
		return fmt.Errorf("%w: %w\n%s", gerrors.ErrPostHookNotExec, err, string(output))
	}

	return nil
}

// isExitError checks if an error is an exec.ExitError.
func (r *Runner) isExitError(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr)
}

// isCommandNotFound checks if an error indicates that a command was not found.
func (r *Runner) isCommandNotFound(err error) bool {
	return strings.Contains(err.Error(), "executable file not found")
}

// isPermissionDenied checks if output contains a permission denied error.
func (r *Runner) isPermissionDenied(output string) bool {
	return strings.Contains(output, "permission denied")
}
