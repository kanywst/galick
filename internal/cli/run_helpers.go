// Package cli provides the command-line interface for galick.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kanywst/galick/internal/config"
	"github.com/kanywst/galick/internal/constants"
	"github.com/kanywst/galick/internal/report"
	"github.com/kanywst/galick/internal/runner"
)

// RunParameters holds the parameters for a scenario run.
type RunParameters struct {
	Scenario    string
	Environment string
	OutputDir   string
}

// prepareRunParameters loads config and determines scenario, environment and output directory.
func (app *App) prepareRunParameters(args []string) (*config.Config, *RunParameters, error) {
	// Load config
	cfg, err := config.LoadConfig(app.Options.CfgFile)
	if err != nil {
		return nil, nil, err
	}

	// Determine scenario to run
	scenarioName := cfg.Default.Scenario
	if len(args) > 0 {
		scenarioName = args[0]
	}

	// Determine environment
	envName := app.Options.Environment
	if envName == "" {
		envName = cfg.Default.Environment
	}

	// Determine output directory
	outDir := app.Options.OutputDir
	if outDir == "" {
		outDir = cfg.Default.OutputDir
	}

	// Create specific output directory for this run
	runOutputDir := filepath.Join(outDir, envName, scenarioName)
	if err := os.MkdirAll(runOutputDir, constants.DirPermissionDefault); err != nil {
		return nil, nil, err
	}

	return cfg, &RunParameters{
		Scenario:    scenarioName,
		Environment: envName,
		OutputDir:   runOutputDir,
	}, nil
}

// executeScenario runs the scenario and generates reports.
func (app *App) executeScenario(r runner.Interface, cfg *config.Config, params *RunParameters) (int, error) {
	_, _ = fmt.Printf("Running scenario '%s' in environment '%s'...\n", params.Scenario, params.Environment)

	// Run the scenario
	result, err := r.RunScenario(cfg, params.Scenario, params.Environment, params.OutputDir)
	if err != nil {
		return 1, err
	}

	// Create reporter
	reporter := report.NewReporter()

	// Generate reports
	_, _ = fmt.Println("Generating reports...")
	results, err := reporter.GenerateReports(
		result.OutputFile,
		params.OutputDir,
		params.Scenario,
		params.Environment,
		cfg,
	)

	if err != nil {
		return 1, err
	}

	// Display report results and check for threshold violations
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

	// Push metrics to Prometheus Pushgateway if configured
	if cfg.Report.Pushgateway.URL != "" {
		app.pushMetricsToGateway(cfg, reporter, result.OutputFile, params.Environment, params.Scenario)
	}

	return exitCode, nil
}

// pushMetricsToGateway extracts metrics and pushes them to Prometheus Pushgateway
func (app *App) pushMetricsToGateway(cfg *config.Config, reporter *report.Reporter, resultFile, environment, scenario string) {
	metrics, metricsErr := reporter.ExtractMetrics(resultFile)
	if metricsErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to extract metrics for Pushgateway: %v\n", metricsErr)
		return
	}

	// Create job name in format galick_<environment>_<scenario>
	jobName := fmt.Sprintf("galick_%s_%s", environment, scenario)

	// Extract Prometheus-formatted metrics
	promMetrics := report.ExtractPrometheusMetrics(metrics)

	// Push metrics to Prometheus
	pushErr := report.PushMetrics(
		cfg.Report.Pushgateway.URL,
		jobName,
		cfg.Report.Pushgateway.Labels,
		promMetrics,
	)

	if pushErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to push metrics to Pushgateway: %v\n", pushErr)
	} else {
		_, _ = fmt.Printf("Metrics pushed to Pushgateway: %s\n", cfg.Report.Pushgateway.URL)
	}
}

// configurePushgateway sets up the Pushgateway configuration from command line arguments and environment variables.
func (app *App) configurePushgateway(cfg *config.Config, pushgatewayURL, pushLabels string) {
	// Override Pushgateway URL from command line if provided
	if pushgatewayURL != "" {
		if cfg.Report.Pushgateway.URL == "" {
			cfg.Report.Pushgateway = config.Pushgateway{
				URL:    pushgatewayURL,
				Labels: make(map[string]string),
			}
		} else {
			cfg.Report.Pushgateway.URL = pushgatewayURL
		}
	}

	// Override with environment variable if set
	if envURL := os.Getenv("GALICK_PUSHGATEWAY_URL"); envURL != "" && cfg.Report.Pushgateway.URL == "" {
		cfg.Report.Pushgateway.URL = envURL
	}

	// Parse additional labels if provided
	if pushLabels != "" {
		labels := parsePushLabels(pushLabels)
		if cfg.Report.Pushgateway.Labels == nil {
			cfg.Report.Pushgateway.Labels = labels
		} else {
			// Merge labels, command-line takes precedence
			for k, v := range labels {
				cfg.Report.Pushgateway.Labels[k] = v
			}
		}
	}
}
