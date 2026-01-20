// Package metrics provides thread-safe statistics collection and reporting.
package metrics

import (
	"sync"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// Result represents the outcome of a single request.
type Result struct {
	Timestamp time.Time
	Latency   time.Duration
	BytesIn   uint64
	BytesOut  uint64
	Code      uint16
	Error     string
}

// Stats holds the aggregated metrics.
type Stats struct {
	TotalRequests uint64
	SuccessCount  uint64
	ErrorCount    uint64
	BytesIn       uint64
	BytesOut      uint64
	
	// Latency Histogram (microsecond resolution)
	// Min: 1us, Max: 1hr, SigFigs: 3
	Histogram *hdrhistogram.Histogram

	mu sync.Mutex
}

// NewStats creates a new Stats instance.
func NewStats() *Stats {
	return &Stats{
		Histogram: hdrhistogram.New(1, int64(time.Hour/time.Microsecond), 3),
	}
}

// Add adds a single request result to the statistics.
func (s *Stats) Add(r Result) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalRequests++
	if r.Error == "" && r.Code >= 200 && r.Code < 400 {
		s.SuccessCount++
	} else {
		s.ErrorCount++
	}

	s.BytesIn += r.BytesIn
	s.BytesOut += r.BytesOut

	if r.Latency > 0 {
		_ = s.Histogram.RecordValue(r.Latency.Microseconds())
	}
}

// Snapshot returns a copy of the current stats for reporting.
func (s *Stats) Snapshot() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clone histogram
	h := hdrhistogram.New(1, int64(time.Hour/time.Microsecond), 3)
	h.Merge(s.Histogram)

	return Stats{
		TotalRequests: s.TotalRequests,
		SuccessCount:  s.SuccessCount,
		ErrorCount:    s.ErrorCount,
		BytesIn:       s.BytesIn,
		BytesOut:      s.BytesOut,
		Histogram:     h,
	}
}

// P99 returns the 99th percentile latency.
func (s *Stats) P99() time.Duration {
	return time.Duration(s.Histogram.ValueAtQuantile(99)) * time.Microsecond
}

// P95 returns the 95th percentile latency.
func (s *Stats) P95() time.Duration {
	return time.Duration(s.Histogram.ValueAtQuantile(95)) * time.Microsecond
}

// Mean returns the mean latency.
func (s *Stats) Mean() time.Duration {
	return time.Duration(s.Histogram.Mean()) * time.Microsecond
}

// Max returns the maximum latency observed.
func (s *Stats) Max() time.Duration {
	return time.Duration(s.Histogram.Max()) * time.Microsecond
}