package metrics

import (
	"testing"
	"time"
)

func TestStats(t *testing.T) {
	s := NewStats()

	// Add some dummy data
	s.Add(Result{
		Latency:  10 * time.Millisecond,
		BytesIn:  100,
		Code:     200,
	})
	s.Add(Result{
		Latency:  20 * time.Millisecond,
		BytesIn:  200,
		Code:     500,
		Error:    "Internal Server Error",
	})

	snap := s.Snapshot()

	if snap.TotalRequests != 2 {
		t.Errorf("Expected 2 requests, got %d", snap.TotalRequests)
	}
	if snap.SuccessCount != 1 {
		t.Errorf("Expected 1 success, got %d", snap.SuccessCount)
	}
	if snap.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", snap.ErrorCount)
	}
	if snap.BytesIn != 300 {
		t.Errorf("Expected 300 bytes in, got %d", snap.BytesIn)
	}

	// Check Latency P50 (should be around 10-20ms)
	// Since we only have 10 and 20, mean is 15.
	if snap.Mean() < 10*time.Millisecond || snap.Mean() > 20*time.Millisecond {
		t.Errorf("Mean latency seems off: %v", snap.Mean())
	}
}
