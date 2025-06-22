// Package runner provides functionality to execute load tests using Vegeta.
package runner

import (
	"github.com/kanywst/galick/internal/config"
)

// Interface defines the interface for load test execution.
type Interface interface {
	RunScenario(cfg *config.Config, scenarioName, environmentName, outputFolder string) (*Result, error)
	RunPreHook(cfg *config.Config) error
	RunPostHook(cfg *config.Config, exitCode int) error
}
