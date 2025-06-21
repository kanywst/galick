// Package config provides configuration loading and validation for galick load tests.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	gerrors "github.com/kanywst/galick/internal/errors"
	"github.com/spf13/viper"
)

// Config represents the main configuration structure for galick.
type Config struct {
	Default      Default                `mapstructure:"default"`
	Environments map[string]Environment `mapstructure:"environments"`
	Scenarios    map[string]Scenario    `mapstructure:"scenarios"`
	Report       Report                 `mapstructure:"report"`
	Hooks        Hooks                  `mapstructure:"hooks"`
}

// Default configuration.
type Default struct {
	Environment string `mapstructure:"environment"`
	Scenario    string `mapstructure:"scenario"`
	OutputDir   string `mapstructure:"output_dir,omitempty"`
}

// Environment configuration.
type Environment struct {
	BaseURL string            `mapstructure:"base_url"`
	Headers map[string]string `mapstructure:"headers"`
}

// Scenario configuration.
type Scenario struct {
	Rate     string   `mapstructure:"rate"`
	Duration string   `mapstructure:"duration"`
	Targets  []string `mapstructure:"targets"`
}

// Report configuration.
type Report struct {
	Formats    []string          `mapstructure:"formats"`
	Thresholds map[string]string `mapstructure:"thresholds"`
}

// Hooks configuration.
type Hooks struct {
	Pre  string `mapstructure:"pre"`
	Post string `mapstructure:"post"`
}

// LoadConfig loads configuration from a file.
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set default values if not specified
	if config.Default.OutputDir == "" {
		config.Default.OutputDir = "output"
	}

	// Initialize map fields if they're nil
	for envName, env := range config.Environments {
		if env.Headers == nil {
			env.Headers = make(map[string]string)
			config.Environments[envName] = env
		}
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// FindAndLoadConfig looks for a loadtest.yaml or loadtest.yml file and loads it.
func FindAndLoadConfig(configPath string) (*Config, error) {
	if configPath != "" {
		return LoadConfig(configPath)
	}

	// Try to find config file in current directory
	for _, name := range []string{"loadtest.yaml", "loadtest.yml"} {
		if _, err := os.Stat(name); err == nil {
			return LoadConfig(name)
		}
	}

	return nil, gerrors.WithConfigNotFoundDetails()
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if err := c.validateDefaults(); err != nil {
		return err
	}
	if err := c.validateEnvironments(); err != nil {
		return err
	}
	if err := c.validateScenarios(); err != nil {
		return err
	}
	return nil
}

// validateDefaults checks if the default settings are valid.
func (c *Config) validateDefaults() error {
	if c.Default.Environment == "" {
		return gerrors.ErrDefaultEnvNotSet
	}
	if c.Default.Scenario == "" {
		return gerrors.ErrDefaultScenarioNotSet
	}
	return nil
}

// validateEnvironments checks if the environments are valid.
func (c *Config) validateEnvironments() error {
	if len(c.Environments) == 0 {
		return gerrors.ErrNoEnvironmentsDefined
	}

	// Check if the default environment exists
	if _, exists := c.Environments[c.Default.Environment]; !exists {
		return gerrors.WithDefaultEnvDetails(c.Default.Environment)
	}

	// Validate each environment
	for name, env := range c.Environments {
		if env.BaseURL == "" {
			return gerrors.WithEnvMissingBaseURLDetails(name)
		}
	}

	return nil
}

// validateScenarios checks if the scenarios are valid.
func (c *Config) validateScenarios() error {
	if len(c.Scenarios) == 0 {
		return gerrors.ErrNoScenariosDefined
	}

	// Check if the default scenario exists
	if _, exists := c.Scenarios[c.Default.Scenario]; !exists {
		return gerrors.WithDefaultScenarioDetails(c.Default.Scenario)
	}

	// Validate each scenario
	for name, scenario := range c.Scenarios {
		if scenario.Rate == "" {
			return gerrors.WithScenarioMissingRateDetails(name)
		}
		if scenario.Duration == "" {
			return gerrors.WithScenarioMissingDurationDetails(name)
		}
		if len(scenario.Targets) == 0 {
			return gerrors.WithScenarioNoTargetsDetails(name)
		}
	}

	return nil
}

// GetOutputPath returns the output path for a specific environment and scenario.
func (c *Config) GetOutputPath(environment, scenario string) string {
	return filepath.Join(c.Default.OutputDir, environment, scenario)
}

// GetEnvironment returns the environment with the given name, or the default environment if empty.
func (c *Config) GetEnvironment(name string) (*Environment, error) {
	if name == "" {
		name = c.Default.Environment
	}

	env, exists := c.Environments[name]
	if !exists {
		return nil, gerrors.WithEnvNotFoundDetails(name)
	}

	return &env, nil
}

// GetScenario returns the scenario with the given name, or the default scenario if empty.
func (c *Config) GetScenario(name string) (*Scenario, error) {
	if name == "" {
		name = c.Default.Scenario
	}

	scenario, exists := c.Scenarios[name]
	if !exists {
		return nil, gerrors.WithScenarioNotFoundDetails(name)
	}

	return &scenario, nil
}
