// Package config provides configuration loading and validation for galick load tests.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

	return nil, errors.New(
		"config file not found: create a loadtest.yaml or loadtest.yml file in the current directory, " +
			"or specify a path with --config",
	)
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Check if default environment and scenario are set
	if c.Default.Environment == "" {
		return errors.New("default environment is not set")
	}
	if c.Default.Scenario == "" {
		return errors.New("default scenario is not set")
	}

	// Check if environments are defined
	if len(c.Environments) == 0 {
		return errors.New("no environments defined in configuration")
	}

	// Check if scenarios are defined
	if len(c.Scenarios) == 0 {
		return errors.New("no scenarios defined in configuration")
	}

	// Check if the default environment exists
	if _, exists := c.Environments[c.Default.Environment]; !exists {
		return fmt.Errorf("default environment '%s' does not exist in the environments section", c.Default.Environment)
	}

	// Check if the default scenario exists
	if _, exists := c.Scenarios[c.Default.Scenario]; !exists {
		return fmt.Errorf("default scenario '%s' does not exist in the scenarios section", c.Default.Scenario)
	}

	// Validate environments
	for name, env := range c.Environments {
		if env.BaseURL == "" {
			return fmt.Errorf("environment '%s' is missing a base_url", name)
		}
	}

	// Validate scenarios
	for name, scenario := range c.Scenarios {
		if scenario.Rate == "" {
			return fmt.Errorf("scenario '%s' is missing a rate", name)
		}
		if scenario.Duration == "" {
			return fmt.Errorf("scenario '%s' is missing a duration", name)
		}
		if len(scenario.Targets) == 0 {
			return fmt.Errorf("scenario '%s' has no targets defined", name)
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
		return nil, fmt.Errorf("environment '%s' not found", name)
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
		return nil, fmt.Errorf("scenario '%s' not found", name)
	}

	return &scenario, nil
}
