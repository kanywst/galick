package engine

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kanywst/galick/pkg/metrics"
)

type MockAttacker struct {
	count int64
}

func (m *MockAttacker) Name() string { return "mock" }
func (m *MockAttacker) Attack(_ context.Context) metrics.Result {
	atomic.AddInt64(&m.count, 1)
	return metrics.Result{
		Latency: time.Millisecond,
		Code:    200,
	}
}

func TestEngineRun(t *testing.T) {
	attacker := &MockAttacker{}
	// Run for 1 second at 10 QPS
	eng := NewEngine(attacker, 2, 10, 1*time.Second)
	
	start := time.Now()
	eng.Run(context.Background())
	duration := time.Since(start)

	stats := eng.Stats()
	
	if stats.TotalRequests == 0 {
		t.Error("Engine did not make any requests")
	}

	// MockAttacker's internal count check (thread-safe check)
	if val := atomic.LoadInt64(&attacker.count); val == 0 {
		t.Error("MockAttacker was not called")
	}

	// Allow some jitter in timing
	if duration < 1*time.Second {
		t.Errorf("Engine finished too early: %v", duration)
	}
}