// Package main provides the CLI entrypoint for Galick.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/kanywst/galick/pkg/engine"
	"github.com/kanywst/galick/pkg/protocols"
	"github.com/kanywst/galick/pkg/protocols/loadhttp"
	"github.com/kanywst/galick/pkg/protocols/script"
	"github.com/kanywst/galick/pkg/report"
)

var (
	// These variables are populated by ldflags during build
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

var (
	targetURL  string
	method     string
	scriptPath string
	qps        int
	workers    int
	duration   time.Duration
	timeout    time.Duration
	headless   bool
	insecure   bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "galick",
		Short: "A modern, high-performance load testing tool",
		Long: `Galick is a scriptable load testing tool designed for precision and real-time observability. 
It supports static target benchmarking and dynamic Starlark-based scenario scripting.

Examples:
  # Static Attack
  galick --url https://api.example.com --qps 50

  # Dynamic Script Attack
  galick --script attack.star --qps 50

  # CI/Docker Mode (No TUI)
  galick --url https://api.example.com --headless`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(_ *cobra.Command, _ []string) {
			var attacker protocols.Attacker
			var err error

			if scriptPath != "" {
				attacker, err = script.NewScriptAttacker(scriptPath, timeout, insecure)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error loading script: %v\n", err)
					os.Exit(1)
				}
			} else {
				if targetURL == "" {
					fmt.Fprintln(os.Stderr, "Error: --url is required unless --script is provided")
					os.Exit(1)
				}
				attacker = loadhttp.NewAttacker(method, targetURL, timeout, insecure)
			}

			// Initialize Engine
			eng := engine.NewEngine(attacker, workers, qps, duration)

			// Context handling
			ctx := context.Background()

			if headless {
				target := targetURL
				if scriptPath != "" {
					target = scriptPath
				}

				fmt.Printf("Starting load test (Headless Mode)...\n")
				fmt.Printf("Target: %s\nQPS: %d\nDuration: %s\nWorkers: %d\n\n", target, qps, duration, workers)

				start := time.Now()
				// Run directly (blocking)
				eng.Run(ctx)

				// Print final report directly
				fmt.Print(report.GenerateTextReport(eng, start))
			} else {
				// Run in background for TUI
				go eng.Run(ctx)

				// Start TUI
				p := tea.NewProgram(report.NewModel(eng, duration))
				if _, err := p.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
					os.Exit(1)
				}
			}
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Galick version %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Built at: %s\n", buildDate)
		},
	}

	rootCmd.AddCommand(versionCmd)

	rootCmd.Flags().StringVarP(&targetURL, "url", "u", "", "Target URL (required for static mode)")
	rootCmd.Flags().StringVarP(&method, "method", "m", "GET", "HTTP Method")
	rootCmd.Flags().StringVarP(&scriptPath, "script", "s", "", "Path to Starlark script (dynamic mode)")
	rootCmd.Flags().IntVarP(&qps, "qps", "q", 50, "Queries Per Second")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 10, "Number of workers")
	rootCmd.Flags().DurationVarP(&duration, "duration", "d", 10*time.Second, "Duration of the test")
	rootCmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout for each request")
	rootCmd.Flags().BoolVar(&headless, "headless", false, "Run without TUI (useful for CI/Docker)")
	rootCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Skip TLS certificate verification")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
		os.Exit(1)
	}
}
