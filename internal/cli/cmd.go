// Package cli provides the command-line interface for galick.
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

// VersionInfo contains version information set at build time.
type VersionInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

// Options contains command-line options and flags.
type Options struct {
	CfgFile     string
	Environment string
	OutputDir   string
	CIMode      bool
}

// App represents the CLI application.
type App struct {
	VersionInfo VersionInfo
	Options     Options
	viper       *viper.Viper
}

// NewApp creates a new CLI application instance.
func NewApp() *App {
	return &App{
		VersionInfo: VersionInfo{
			Version:   "dev",
			Commit:    "none",
			BuildDate: "unknown",
		},
		Options: Options{},
		viper:   viper.New(),
	}
}

// Run is the main entry point for the galick CLI.
func Run() {
	app := NewApp()
	if err := app.NewRootCmd().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err) // エラー出力は無視
		os.Exit(1)
	}
}

// NewRootCmd creates the root command with all subcommands.
func (app *App) NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "galick",
		Short: "A Vegeta-based load testing tool with enhanced features",
		Long: `Galick - A powerful wrapper around Vegeta for load testing.
Named after Vegeta's most powerful attack, Galick provides extra 
firepower for your load testing workflows.

Complete documentation is available at https://github.com/kanywst/galick`,
	}

	// Persistent flags for the root command
	rootCmd.PersistentFlags().StringVar(&app.Options.CfgFile, "config", "", "config file (default is ./loadtest.yaml)")
	rootCmd.PersistentFlags().StringVarP(&app.Options.Environment, "env", "e", "", "environment to use (default is from config)")
	rootCmd.PersistentFlags().StringVarP(&app.Options.OutputDir, "output-dir", "o", "", "output directory (default is from config)")
	rootCmd.PersistentFlags().BoolVar(
		&app.Options.CIMode,
		"ci",
		false,
		"enable CI mode (exit with non-zero code on threshold violations)",
	)

	// Add subcommands
	rootCmd.AddCommand(app.newVersionCmd())
	rootCmd.AddCommand(app.newInitCmd())
	rootCmd.AddCommand(app.newRunCmd())
	rootCmd.AddCommand(app.newReportCmd())

	// Initialize config
	app.initConfig()

	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func (app *App) initConfig() {
	if app.Options.CfgFile != "" {
		// Use config file from the flag
		app.viper.SetConfigFile(app.Options.CfgFile)
	} else {
		// Use default config location
		app.viper.AddConfigPath(".")
		app.viper.SetConfigName("loadtest")
		app.viper.SetConfigType("yaml")
	}

	app.viper.AutomaticEnv()

	// If a config file is found, read it in
	_ = app.viper.ReadInConfig() // エラーは無視する
}

// newVersionCmd creates the version command.
func (app *App) newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintf(
				cmd.OutOrStdout(),
				"galick version %s (commit: %s, built: %s)\n",
				app.VersionInfo.Version, app.VersionInfo.Commit, app.VersionInfo.BuildDate,
			)
		},
	}
}

// newInitCmd creates the init command.
func (app *App) newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new load test configuration",
		Long:  `Creates a new loadtest.yaml file with default settings.`,
		Run: func(cmd *cobra.Command, _ []string) {
			// Check if config file already exists
			configFile := "loadtest.yaml"
			if app.Options.CfgFile != "" {
				configFile = app.Options.CfgFile
			}

			if _, err := os.Stat(configFile); err == nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Config file %s already exists. Overwrite? (y/n): ", configFile)
				var response string
				_, err := fmt.Scanln(&response)
				if err != nil {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Error reading input:", err)
					return
				}
				if strings.ToLower(response) != "y" {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
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
			if err := os.WriteFile(configFile, []byte(defaultConfig), 0o600); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Error creating config file:", err)
				os.Exit(1)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created config file: %s\n", configFile)
		},
	}
}

// newRunCmd creates the run command.
func (app *App) newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [scenario]",
		Short: "Run a load test scenario",
		Args:  cobra.MaximumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			// Load config and prepare parameters
			cfg, runParams, err := app.prepareRunParameters(args)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			// Create runner
			r := runner.NewRunner()

			// Run pre-hook if configured
			if err := r.RunPreHook(cfg); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Pre-hook error:", err)
				os.Exit(1)
			}

			// Run the scenario
			exitCode, err := app.executeScenario(r, cfg, runParams)

			// Run post-hook
			if hookErr := r.RunPostHook(cfg, exitCode); hookErr != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Post-hook error:", hookErr)
				os.Exit(1)
			}

			// Exit if there was an error or threshold violations in CI mode
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}

	return cmd
}

// newReportCmd creates the report command.
func (app *App) newReportCmd() *cobra.Command {
	var (
		resultsFile string
		reportDir   string
		reportType  string
	)

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a report from existing results",
		Run: func(_ *cobra.Command, _ []string) {
			// Validate flags
			if resultsFile == "" {
				_, _ = fmt.Fprintln(os.Stderr, "Error: results file is required")
				os.Exit(1)
			}

			// Load config
			cfg, err := config.LoadConfig(app.Options.CfgFile)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			// Determine report directory
			if reportDir == "" {
				reportDir = filepath.Dir(resultsFile)
			}

			// Create reporter
			reporter := report.NewReporter()

			// Generate report based on type
			exitCode := app.generateReport(reporter, cfg, resultsFile, reportDir, reportType)

			// Exit with appropriate code in CI mode
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}

	cmd.Flags().StringVarP(&resultsFile, "results", "r", "", "path to the results file")
	cmd.Flags().StringVarP(
		&reportDir,
		"dir",
		"d",
		"",
		"directory to save reports (defaults to same directory as results)",
	)
	cmd.Flags().StringVarP(&reportType, "type", "t", "", "report type (json, text, markdown, html)")

	return cmd
}

// generateReport handles the report generation logic based on the specified type.
// Returns an exit code (0 for success, 1 for threshold violations in CI mode)
func (app *App) generateReport(
	reporter *report.Reporter,
	cfg *config.Config,
	resultsFile,
	reportDir,
	reportType string,
) int {
	// Generate specific report type if requested
	if reportType != "" {
		return app.generateSingleReport(reporter, resultsFile, reportDir, reportType)
	}

	// Otherwise generate all configured report formats
	return app.generateAllReports(reporter, cfg, resultsFile, reportDir)
}

// generateSingleReport generates a single report of the specified type.
// Returns an exit code (always 0 for single reports as thresholds aren't checked)
func (app *App) generateSingleReport(
	reporter *report.Reporter,
	resultsFile,
	reportDir,
	reportType string,
) int {
	_, _ = fmt.Printf("Generating %s report...\n", reportType)
	outputFile := filepath.Join(reportDir, fmt.Sprintf("report.%s", report.FormatExtension(reportType)))

	// Extract metrics for threshold validation
	metrics, err := reporter.ExtractMetrics(resultsFile)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	err = reporter.GenerateReport(resultsFile, outputFile, reportType, metrics)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	_, _ = fmt.Printf("Report saved to: %s\n", outputFile)
	return 0
}

// generateAllReports generates all configured report formats.
// Returns an exit code (0 for success, 1 for threshold violations in CI mode)
func (app *App) generateAllReports(
	reporter *report.Reporter,
	cfg *config.Config,
	resultsFile,
	reportDir string,
) int {
	_, _ = fmt.Println("Generating reports...")
	results, err := reporter.GenerateReports(
		resultsFile,
		reportDir,
		"", // scenario name not used for standalone reports
		"", // environment name not used for standalone reports
		cfg,
	)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	// Display report results
	exitCode := 0
	for _, report := range results {
		if !report.Passed {
			_, _ = fmt.Printf("⚠️ Threshold violations detected in %s report\n", report.Format)
			if app.Options.CIMode {
				exitCode = 1
			}
		}
		_, _ = fmt.Printf("Report saved to: %s\n", report.FilePath)
	}

	return exitCode
}
