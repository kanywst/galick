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
func (app *App) executeScenario(r *runner.Runner, cfg *config.Config, params *RunParameters) (int, error) {
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

	return exitCode, nil
}
