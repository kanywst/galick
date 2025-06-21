// Package cli provides the command-line interface for galick
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kanywst/galick/internal/config"
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

// Run is the main entry point for the galick CLI
func Run() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// NewRootCmd creates the root command with all subcommands
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "galick",
		Short: "A Vegeta-based load testing tool with enhanced features",
		Long: `Galick - A powerful wrapper around Vegeta for load testing.
Named after Vegeta's most powerful attack, Galick provides extra 
firepower for your load testing workflows.

Complete documentation is available at https://github.com/kanywst/galick`,
	}

	// Persistent flags for the root command
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./loadtest.yaml)")
	rootCmd.PersistentFlags().StringVarP(&environment, "env", "e", "", "environment to use (default is from config)")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output-dir", "o", "", "output directory (default is from config)")
	rootCmd.PersistentFlags().BoolVar(&ciMode, "ci", false, "enable CI mode (exit with non-zero code on threshold violations)")

	// Add subcommands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newReportCmd())

	// Initialize config
	cobra.OnInitialize(initConfig)

	return rootCmd
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Use default config location
		viper.AddConfigPath(".")
		viper.SetConfigName("loadtest")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "galick version %s (commit: %s, built: %s)\n", version, commit, buildDate)
		},
	}
}

// newInitCmd creates the init command
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new load test configuration",
		Long:  `Creates a new loadtest.yaml file with default settings.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Check if config file already exists
			configFile := "loadtest.yaml"
			if cfgFile != "" {
				configFile = cfgFile
			}

			if _, err := os.Stat(configFile); err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Config file %s already exists. Overwrite? (y/n): ", configFile)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return
				}
			}

			// Create default config
			defaultConfig := `# galick load test configuration

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

			// Write the config file
			if err := os.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
				fmt.Fprintln(os.Stderr, "Error creating config file:", err)
				os.Exit(1)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created config file: %s\n", configFile)
		},
	}
}

// newRunCmd creates the run command
func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [scenario]",
		Short: "Run a load test scenario",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Load config
			cfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			// Determine scenario to run
			scenarioName := cfg.Default.Scenario
			if len(args) > 0 {
				scenarioName = args[0]
			}

			// Determine environment
			envName := environment
			if envName == "" {
				envName = cfg.Default.Environment
			}

			// Determine output directory
			outDir := outputDir
			if outDir == "" {
				outDir = cfg.Default.OutputDir
			}

			// Create specific output directory for this run
			runOutputDir := filepath.Join(outDir, envName, scenarioName)
			if err := os.MkdirAll(runOutputDir, 0755); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			// Create runner
			r := runner.NewRunner()

			// Run pre-hook if configured
			if err := r.RunPreHook(cfg); err != nil {
				fmt.Fprintln(os.Stderr, "Pre-hook error:", err)
				os.Exit(1)
			}

			// Run the scenario
			fmt.Printf("Running scenario '%s' in environment '%s'...\n", scenarioName, envName)
			result, err := r.RunScenario(cfg, scenarioName, envName, runOutputDir)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				// Run post-hook with error status
				if hookErr := r.RunPostHook(cfg, 1); hookErr != nil {
					fmt.Fprintln(os.Stderr, "Post-hook error:", hookErr)
				}
				os.Exit(1)
			}

			// Create reporter
			reporter := report.NewReporter()

			// Generate reports
			fmt.Println("Generating reports...")
			reportResults, err := reporter.GenerateReports(
				result.OutputFile,
				runOutputDir,
				scenarioName,
				envName,
				cfg,
			)

			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				// Run post-hook with error status
				if hookErr := r.RunPostHook(cfg, 1); hookErr != nil {
					fmt.Fprintln(os.Stderr, "Post-hook error:", hookErr)
				}
				os.Exit(1)
			}

			// Display report results
			exitCode := 0
			for _, report := range reportResults {
				if !report.Passed {
					fmt.Printf("⚠️ Threshold violations detected in %s report\n", report.Format)
					if ciMode {
						exitCode = 1
					}
				}
				fmt.Printf("Report saved to: %s\n", report.FilePath)
			}

			// Run post-hook
			if err := r.RunPostHook(cfg, exitCode); err != nil {
				fmt.Fprintln(os.Stderr, "Post-hook error:", err)
				os.Exit(1)
			}

			// Exit with appropriate code in CI mode
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}

	return cmd
}

// newReportCmd creates the report command
func newReportCmd() *cobra.Command {
	var (
		resultsFile string
		reportDir   string
		reportType  string
	)

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a report from existing results",
		Run: func(cmd *cobra.Command, args []string) {
			// Validate flags
			if resultsFile == "" {
				fmt.Fprintln(os.Stderr, "Error: results file is required")
				os.Exit(1)
			}

			// Load config
			cfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			// Determine report directory
			if reportDir == "" {
				reportDir = filepath.Dir(resultsFile)
			}

			// Create reporter
			reporter := report.NewReporter()

			// Generate specific report type if requested
			if reportType != "" {
				fmt.Printf("Generating %s report...\n", reportType)
				outputFile := filepath.Join(reportDir, fmt.Sprintf("report.%s", report.FormatExtension(reportType)))

				// Extract metrics for threshold validation
				metrics, err := reporter.ExtractMetrics(resultsFile)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					os.Exit(1)
				}

				err = reporter.GenerateReport(resultsFile, outputFile, reportType, metrics)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					os.Exit(1)
				}

				fmt.Printf("Report saved to: %s\n", outputFile)
				return
			}

			// Otherwise generate all configured report formats
			fmt.Println("Generating reports...")
			reportResults, err := reporter.GenerateReports(
				resultsFile,
				reportDir,
				"", // scenario name not used for standalone reports
				"", // environment name not used for standalone reports
				cfg,
			)

			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			// Display report results
			exitCode := 0
			for _, report := range reportResults {
				if !report.Passed {
					fmt.Printf("⚠️ Threshold violations detected in %s report\n", report.Format)
					if ciMode {
						exitCode = 1
					}
				}
				fmt.Printf("Report saved to: %s\n", report.FilePath)
			}

			// Exit with appropriate code in CI mode
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}

	cmd.Flags().StringVarP(&resultsFile, "results", "r", "", "path to the results file")
	cmd.Flags().StringVarP(&reportDir, "dir", "d", "", "directory to save reports (defaults to same directory as results)")
	cmd.Flags().StringVarP(&reportType, "type", "t", "", "report type (json, text, markdown, html)")

	return cmd
}
