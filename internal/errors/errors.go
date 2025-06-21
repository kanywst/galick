// Package errors provides custom error types for galick.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// Common errors.
var (
	// Configuration errors
	ErrConfigNotFound          = errors.New("config file not found")
	ErrDefaultEnvNotSet        = errors.New("default environment is not set")
	ErrDefaultScenarioNotSet   = errors.New("default scenario is not set")
	ErrNoEnvironmentsDefined   = errors.New("no environments defined in configuration")
	ErrNoScenariosDefined      = errors.New("no scenarios defined in configuration")
	ErrDefaultEnvNotExist      = errors.New("default environment does not exist in the environments section")
	ErrDefaultScenarioNotExist = errors.New("default scenario does not exist in the scenarios section")
	ErrEnvMissingBaseURL       = errors.New("environment is missing a base_url")
	ErrScenarioMissingRate     = errors.New("scenario is missing a rate")
	ErrScenarioMissingDuration = errors.New("scenario is missing a duration")
	ErrScenarioNoTargets       = errors.New("scenario has no targets defined")
	ErrEnvironmentNotFound     = errors.New("environment not found")
	ErrScenarioNotFound        = errors.New("scenario not found")
	ErrConfigNil               = errors.New("config is nil")

	// Hook errors
	ErrPreHookNotFound  = errors.New("pre-hook script not found")
	ErrPreHookNotExec   = errors.New("pre-hook script is not executable")
	ErrPostHookNotFound = errors.New("post-hook script not found")
	ErrPostHookNotExec  = errors.New("post-hook script is not executable")

	// Report errors
	ErrMetricsNil         = errors.New("metrics are nil")
	ErrResultFileEmpty    = errors.New("result file path is empty")
	ErrOutputDirEmpty     = errors.New("output directory path is empty")
	ErrResultFileNotFound = errors.New("result file not found")
	ErrThresholdViolation = errors.New("threshold violations")

	// Latency errors
	ErrInvalidMsValue     = errors.New("invalid value for milliseconds")
	ErrInvalidSecValue    = errors.New("invalid value for seconds")
	ErrUnknownLatencyUnit = errors.New("unknown latency unit in threshold")

	// Vegeta errors
	ErrVegetaAttackFailed = errors.New("vegeta attack failed")
	ErrVegetaNotFound     = errors.New("vegeta command not found")
	ErrVegetaNotExec      = errors.New("permission denied when executing vegeta")
	ErrOutputNotCreated   = errors.New("output file was not created")

	// Target errors
	ErrEnvironmentBaseURLEmpty = errors.New("environment base URL is empty")
	ErrInvalidTargetFormat     = errors.New("invalid target format")
	ErrInvalidHTTPMethod       = errors.New("invalid HTTP method")
)

// WithConfigNotFoundDetails adds details to the config not found error.
func WithConfigNotFoundDetails() error {
	return fmt.Errorf("%w: create a loadtest.yaml or loadtest.yml file in the current directory, or specify a path with --config", ErrConfigNotFound)
}

// WithDefaultEnvDetails adds the environment name to the default environment error.
func WithDefaultEnvDetails(name string) error {
	return fmt.Errorf("%w: '%s'", ErrDefaultEnvNotExist, name)
}

// WithDefaultScenarioDetails adds the scenario name to the default scenario error.
func WithDefaultScenarioDetails(name string) error {
	return fmt.Errorf("%w: '%s'", ErrDefaultScenarioNotExist, name)
}

// WithEnvMissingBaseURLDetails adds the environment name to the missing base URL error.
func WithEnvMissingBaseURLDetails(name string) error {
	return fmt.Errorf("%w: '%s'", ErrEnvMissingBaseURL, name)
}

// WithScenarioMissingRateDetails adds the scenario name to the missing rate error.
func WithScenarioMissingRateDetails(name string) error {
	return fmt.Errorf("%w: '%s'", ErrScenarioMissingRate, name)
}

// WithScenarioMissingDurationDetails adds the scenario name to the missing duration error.
func WithScenarioMissingDurationDetails(name string) error {
	return fmt.Errorf("%w: '%s'", ErrScenarioMissingDuration, name)
}

// WithScenarioNoTargetsDetails adds the scenario name to the no targets error.
func WithScenarioNoTargetsDetails(name string) error {
	return fmt.Errorf("%w: '%s'", ErrScenarioNoTargets, name)
}

// WithEnvNotFoundDetails adds the environment name to the environment not found error.
func WithEnvNotFoundDetails(name string) error {
	return fmt.Errorf("%w: '%s'", ErrEnvironmentNotFound, name)
}

// WithScenarioNotFoundDetails adds the scenario name to the scenario not found error.
func WithScenarioNotFoundDetails(name string) error {
	return fmt.Errorf("%w: '%s'", ErrScenarioNotFound, name)
}

// WithPreHookNotFoundDetails adds the hook path to the pre-hook not found error.
func WithPreHookNotFoundDetails(path string) error {
	return fmt.Errorf("%w: %s", ErrPreHookNotFound, path)
}

// WithPreHookNotExecDetails adds the hook path to the pre-hook not executable error.
func WithPreHookNotExecDetails(path string) error {
	return fmt.Errorf("%w: %s", ErrPreHookNotExec, path)
}

// WithPostHookNotFoundDetails adds the hook path to the post-hook not found error.
func WithPostHookNotFoundDetails(path string) error {
	return fmt.Errorf("%w: %s", ErrPostHookNotFound, path)
}

// WithPostHookNotExecDetails adds the hook path to the post-hook not executable error.
func WithPostHookNotExecDetails(path string) error {
	return fmt.Errorf("%w: %s", ErrPostHookNotExec, path)
}

// WithResultFileNotFoundDetails adds the file path to the result file not found error.
func WithResultFileNotFoundDetails(path string) error {
	return fmt.Errorf("%w: %s", ErrResultFileNotFound, path)
}

// WithThresholdViolationDetails adds the violations to the threshold violation error.
func WithThresholdViolationDetails(violations []string) error {
	return fmt.Errorf("%w: %s", ErrThresholdViolation, strings.Join(violations, ", "))
}

// WithInvalidMsValueDetails adds the threshold to the invalid milliseconds error.
func WithInvalidMsValueDetails(threshold string) error {
	return fmt.Errorf("%w: %s", ErrInvalidMsValue, threshold)
}

// WithInvalidSecValueDetails adds the threshold to the invalid seconds error.
func WithInvalidSecValueDetails(threshold string) error {
	return fmt.Errorf("%w: %s", ErrInvalidSecValue, threshold)
}

// WithUnknownLatencyUnitDetails adds the threshold to the unknown latency unit error.
func WithUnknownLatencyUnitDetails(threshold string) error {
	return fmt.Errorf("%w: %s", ErrUnknownLatencyUnit, threshold)
}

// WithVegetaAttackFailedDetails adds the exit code and output to the vegeta attack failed error.
func WithVegetaAttackFailedDetails(exitCode int, output string) error {
	return fmt.Errorf("%w with exit code %d: %s", ErrVegetaAttackFailed, exitCode, output)
}

// WithVegetaNotFoundDetails adds additional info to the vegeta not found error.
func WithVegetaNotFoundDetails() error {
	return fmt.Errorf("%w. Please install Vegeta (https://github.com/tsenart/vegeta) and make sure it's in your PATH", ErrVegetaNotFound)
}

// WithInvalidTargetFormatDetails adds the target index and value to the invalid target format error.
func WithInvalidTargetFormatDetails(index int, target string) error {
	return fmt.Errorf("%w at index %d: %s (expected 'METHOD /path')", ErrInvalidTargetFormat, index, target)
}

// WithInvalidHTTPMethodDetails adds the method index and value to the invalid HTTP method error.
func WithInvalidHTTPMethodDetails(index int, method string) error {
	return fmt.Errorf("%w at index %d: %s", ErrInvalidHTTPMethod, index, method)
}
