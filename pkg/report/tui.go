// Package report provides reporting functionality, including a TUI and text reports.
package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kanywst/galick/pkg/engine"
)

// TickMsg is a message sent on each tick to update the UI.
type TickMsg time.Time

// Model represents the state of the TUI.
type Model struct {
	engine   *engine.Engine
	progress progress.Model
	quitting bool
	duration time.Duration
	start    time.Time
}

// NewModel creates a new TUI model.
func NewModel(eng *engine.Engine, duration time.Duration) Model {
	return Model{
		engine:   eng,
		duration: duration,
		start:    time.Now(),
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

// Init initializes the TUI.
func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update handles UI messages and updates the state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	case TickMsg:
		if m.quitting {
			return m, tea.Quit
		}
		
		elapsed := time.Since(m.start)
		if elapsed >= m.duration {
			m.quitting = true
			return m, tea.Quit
		}

		// Update progress
		percent := float64(elapsed) / float64(m.duration)
		if percent > 1.0 {
			percent = 1.0
		}
		
		// Return new tick
		return m, tea.Batch(
			m.progress.SetPercent(percent),
			tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
				return TickMsg(t)
			}),
		)
		
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}
	return m, nil
}

// View renders the UI.
func (m Model) View() string {
	if m.quitting {
		return m.FinalReport()
	}

	stats := m.engine.Stats()
	elapsed := time.Since(m.start).Seconds()
	currentQPS := float64(stats.TotalRequests) / elapsed

	s := strings.Builder{}
	s.WriteString("\n  Galick Load Test Running...\n\n")
	
	// Stats Grid
	s.WriteString(fmt.Sprintf("  Requests: %d\n", stats.TotalRequests))
	s.WriteString(fmt.Sprintf("  Success:  %d\n", stats.SuccessCount))
	s.WriteString(fmt.Sprintf("  Errors:   %d\n", stats.ErrorCount))
	s.WriteString(fmt.Sprintf("  QPS:      %.2f\n", currentQPS))
	s.WriteString(fmt.Sprintf("  P99:      %v\n", stats.P99()))
	s.WriteString("\n")
	
	s.WriteString("  " + m.progress.View() + "\n")
	s.WriteString("\n  Press 'q' to quit\n")

	return s.String()
}

// FinalReport returns the final summary report as a string.
func (m Model) FinalReport() string {
	return GenerateTextReport(m.engine, m.start)
}

// GenerateTextReport creates the final summary string.
func GenerateTextReport(eng *engine.Engine, start time.Time) string {
	stats := eng.Stats()
	elapsed := time.Since(start)
	
	elapsedSec := elapsed.Seconds()
	avgQPS := float64(stats.TotalRequests) / elapsedSec

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(1)

	s := strings.Builder{}
	s.WriteString("\n")
	s.WriteString(style.Render(" TEST COMPLETED "))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("  Duration:   %v\n", elapsed))
	s.WriteString(fmt.Sprintf("  Requests:   %d\n", stats.TotalRequests))
	s.WriteString(fmt.Sprintf("  Mean QPS:   %.2f\n", avgQPS))
	if stats.TotalRequests > 0 {
		s.WriteString(fmt.Sprintf("  Success:    %.2f%%\n", float64(stats.SuccessCount)/float64(stats.TotalRequests)*100))
	} else {
		s.WriteString("  Success:    0.00%\n")
	}
	
	if stats.TotalRequests > 0 {
		s.WriteString(fmt.Sprintf("  P50 Latency: %v\n", time.Duration(stats.Histogram.ValueAtQuantile(50))*time.Microsecond))
		s.WriteString(fmt.Sprintf("  P95 Latency: %v\n", stats.P95()))
		s.WriteString(fmt.Sprintf("  P99 Latency: %v\n", stats.P99()))
		s.WriteString(fmt.Sprintf("  Max Latency: %v\n", stats.Max()))
	}
	s.WriteString("\n")

	return s.String()
}