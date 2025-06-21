// Package cmd provides the command-line interface for galick
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kanywst/galick/internal/config"
	"github.com/kanywst/galick/internal/hooks"
	"github.com/kanywst/galick/internal/report"
	"github.com/kanywst/galick/internal/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version information (set at build time)
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"

	// CLI flags
	cfgFile     string
	environment string
	outputDir   string
	formats     []string
	ciMode      bool
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// newRootCmd creates a new root command for galick
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "galick",
		Short: "galick is a load testing tool powered by Vegeta",
		Long: `galick is a load testing tool that provides enhanced features over Vegeta.
It allows you to define scenarios and environments in a YAML file, 
generate different report formats, and run hooks before and after tests.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Check if CI mode is enabled via environment variable
			if os.Getenv("CI") == "true" {
				ciMode = true
			}
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./loadtest.yaml)")
	rootCmd.PersistentFlags().StringVarP(&environment, "env", "e", "", "environment to use (default is from config)")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output-dir", "o", "", "output directory (default is from config)")
	rootCmd.PersistentFlags().BoolVar(&ciMode, "ci", false, "enable CI mode (exit with non-zero code on threshold violations)")

	// Add commands
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newReportCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

// newInitCmd creates a new init command
func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new loadtest.yaml configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd)
		},
	}

	return cmd
}

// newRunCmd creates a new run command
func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [scenario]",
		Short: "Run a load test scenario",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var scenario string
			if len(args) > 0 {
				scenario = args[0]
			}
			return runScenario(cmd, scenario)
		},
	}

	return cmd
}

// newReportCmd creates a new report command
func newReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report [scenario]",
		Short: "Generate reports for a scenario",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var scenario string
			if len(args) > 0 {
				scenario = args[0]
			}
			return generateReport(cmd, scenario)
		},
	}

	cmd.Flags().StringSliceVarP(&formats, "format", "f", nil, "report formats to generate (json, html, markdown)")

	return cmd
}

// newVersionCmd creates a new version command
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "galick version %s (commit: %s, built: %s)\n", version, commit, buildDate)
		},
	}

	return cmd
}

// runInit initializes a new configuration file
func runInit(cmd *cobra.Command) error {
	// Check if file already exists
	configFile := "loadtest.yaml"
	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("configuration file %s already exists", configFile)
	}

	// Sample configuration
	sampleConfig := `# galick load test configuration

default:
  environment: dev
  scenario: simple
  output_dir: output

environments:
  dev:
    base_url: http://localhost:8080
    headers:
      Content-Type: application/json
  staging:
    base_url: https://staging.example.com
    headers:
      Content-Type: application/json
      Authorization: Bearer YOUR_TOKEN

scenarios:
  simple:
    rate: 10/s
    duration: 30s
    targets:
      - GET /api/health
      - GET /api/users
  heavy:
    rate: 50/s
    duration: 60s
    targets:
      - GET /api/products
      - POST /api/orders

report:
  formats:
    - json
    - markdown
  thresholds:
    p95: 200ms
    success_rate: 99.0

hooks:
  pre: ./scripts/pre-load.sh
  post: ./scripts/post-load.sh
`

	// Write the file
	if err := os.WriteFile(configFile, []byte(sampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to create configuration file: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created %s with sample configuration\n", configFile)
	return nil
}

// runScenario runs a load test scenario
func runScenario(cmd *cobra.Command, scenarioName string) error {
	// Load configuration
	cfg, err := loadConfig(cfgFile)
	if err != nil {
		return err
	}

	// Override environment if specified
	env := environment
	if env == "" {
		env = cfg.Default.Environment
	}

	// Override output directory if specified
	outDir := outputDir
	if outDir == "" {
		outDir = cfg.Default.OutputDir
	}

	// Create output directory path
	outputPath := filepath.Join(outDir, env, scenarioName)
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create runners
	r := runner.NewRunner()
	h := hooks.NewHookRunner()

	// Run pre-hook
	if err := h.RunPreHook(cfg); err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Warning: pre-hook failed: %v\n", err)
	}

	// Run the scenario
	fmt.Fprintf(cmd.OutOrStdout(), "Running scenario '%s' in environment '%s'...\n", scenarioName, env)
	result, err := r.RunScenario(cfg, scenarioName, env, outputPath)
	if err != nil {
		return fmt.Errorf("failed to run scenario: %w", err)
	}

	// Generate reports
	fmt.Fprintf(cmd.OutOrStdout(), "Generating reports...\n")
	reporter := report.NewReporter()
	reportResults, err := reporter.GenerateReports(
		result.OutputFile,
		outputPath,
		scenarioName,
		env,
		cfg,
	)
	if err != nil {
		return fmt.Errorf("failed to generate reports: %w", err)
	}

	// Check if any thresholds were violated
	exitCode := 0
	thresholdViolations := false
	for _, res := range reportResults {
		if !res.Passed {
			thresholdViolations = true
			fmt.Fprintf(cmd.OutOrStdout(), "⚠️ Threshold violations detected in %s report\n", res.Format)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Report saved to: %s\n", res.FilePath)
	}

	// Set exit code for CI mode
	if thresholdViolations && ciMode {
		exitCode = 1
		fmt.Fprintf(cmd.OutOrStdout(), "CI mode enabled: exiting with code 1 due to threshold violations\n")
	}

	// Run post-hook
	if err := h.RunPostHook(cfg, exitCode); err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Warning: post-hook failed: %v\n", err)
	}

	// Set exit code for CI mode
	if thresholdViolations && ciMode {
		os.Exit(1)
	}

	return nil
}

// generateReport generates reports for a scenario
func generateReport(cmd *cobra.Command, scenarioName string) error {
	// Load configuration
	cfg, err := loadConfig(cfgFile)
	if err != nil {
		return err
	}

	// Override environment if specified
	env := environment
	if env == "" {
		env = cfg.Default.Environment
	}

	// Override output directory if specified
	outDir := outputDir
	if outDir == "" {
		outDir = cfg.Default.OutputDir
	}

	// If scenarioName is empty, use default
	if scenarioName == "" {
		scenarioName = cfg.Default.Scenario
	}

	// Override formats if specified
	if len(formats) > 0 {
		cfg.Report.Formats = formats
	}

	// Find the results file
	outputPath := filepath.Join(outDir, env, scenarioName)
	resultsFile := filepath.Join(outputPath, "results.bin")

	if _, err := os.Stat(resultsFile); os.IsNotExist(err) {
		return fmt.Errorf("results file not found for scenario '%s' in environment '%s'", scenarioName, env)
	}

	// Generate reports
	fmt.Fprintf(cmd.OutOrStdout(), "Generating reports for scenario '%s' in environment '%s'...\n", scenarioName, env)
	reporter := report.NewReporter()
	reportResults, err := reporter.GenerateReports(
		resultsFile,
		outputPath,
		scenarioName,
		env,
		cfg,
	)
	if err != nil {
		return fmt.Errorf("failed to generate reports: %w", err)
	}

	// Check if any thresholds were violated
	exitCode := 0
	thresholdViolations := false
	for _, res := range reportResults {
		if !res.Passed {
			thresholdViolations = true
			fmt.Fprintf(cmd.OutOrStdout(), "⚠️ Threshold violations detected in %s report\n", res.Format)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Report saved to: %s\n", res.FilePath)
	}

	// Set exit code for CI mode
	if thresholdViolations && ciMode {
		exitCode = 1
		fmt.Fprintf(cmd.OutOrStdout(), "CI mode enabled: exiting with code 1 due to threshold violations\n")
	}

	// Run hooks
	h := hooks.NewHookRunner()
	if err := h.RunPostHook(cfg, exitCode); err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Warning: post-hook failed: %v\n", err)
	}

	// Set exit code for CI mode
	if thresholdViolations && ciMode {
		os.Exit(1)
	}

	return nil
}

// loadConfig loads the configuration from file
func loadConfig(configPath string) (*config.Config, error) {
	// Initialize viper
	v := viper.New()
	
	// Set default config name and paths
	v.SetConfigName("loadtest")
	v.AddConfigPath(".")
	
	// Set environment variables prefix
	v.SetEnvPrefix("GALICK")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// If a config file is provided, use it
	if configPath != "" {
		v.SetConfigFile(configPath)
	}
	
	// Load the configuration
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Set default values if not specified
	if cfg.Default.OutputDir == "" {
		cfg.Default.OutputDir = "output"
	}
	
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	
	return &cfg, nil
}
