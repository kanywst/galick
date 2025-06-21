// Package cli provides the command-line interface for galick.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kanywst/galick/internal/config"
	"github.com/kanywst/galick/internal/report"
	"github.com/kanywst/galick/internal/runner"
)

// prepareRunParameters loads config and determines scenario, environment and output directory.
func prepareRunParameters(args []string) (*config.Config, *runParameters, error) {
	// Load config
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return nil, nil, err
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
	if err := os.MkdirAll(runOutputDir, 0o755); err != nil {
		return nil, nil, err
	}

	return cfg, &runParameters{
		scenarioName: scenarioName,
		envName:      envName,
		outputDir:    runOutputDir,
	}, nil
}

// runParameters holds the parameters for a scenario run.
type runParameters struct {
	scenarioName string
	envName      string
	outputDir    string
}

// executeScenario runs the scenario and generates reports.
func executeScenario(r *runner.Runner, cfg *config.Config, params *runParameters) (int, error) {
	_, _ = fmt.Printf("Running scenario '%s' in environment '%s'...\n", params.scenarioName, params.envName)

	// Run the scenario
	result, err := r.RunScenario(cfg, params.scenarioName, params.envName, params.outputDir)
	if err != nil {
		return 1, err
	}

	// Create reporter
	reporter := report.NewReporter()

	// Generate reports
	_, _ = fmt.Println("Generating reports...")
	reportResults, err := reporter.GenerateReports(
		result.OutputFile,
		params.outputDir,
		params.scenarioName,
		params.envName,
		cfg,
	)

	if err != nil {
		return 1, err
	}

	// Display report results and check for threshold violations
	exitCode := 0
	for _, report := range reportResults {
		if !report.Passed {
			_, _ = fmt.Printf("⚠️ Threshold violations detected in %s report\n", report.Format)
			if ciMode {
				exitCode = 1
			}
		}
		_, _ = fmt.Printf("Report saved to: %s\n", report.FilePath)
	}

	return exitCode, nil
}
