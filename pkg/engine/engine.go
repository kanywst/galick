// Package engine provides the core load testing execution logic.
package engine

import (
	"context"
	"sync"
	"time"

	"github.com/kanywst/galick/pkg/metrics"
	"github.com/kanywst/galick/pkg/protocols"
)

// Engine manages the load testing execution.
type Engine struct {
	attacker protocols.Attacker
	stats    *metrics.Stats
	workers  int
	qps      int
	duration time.Duration
}

// NewEngine creates a new load testing engine.
func NewEngine(attacker protocols.Attacker, workers, qps int, duration time.Duration) *Engine {
	return &Engine{
		attacker: attacker,
		stats:    metrics.NewStats(),
		workers:  workers,
		qps:      qps,
		duration: duration,
	}
}

// Stats returns the current statistics of the engine.
func (e *Engine) Stats() *metrics.Stats {
	return e.stats
}

// Run starts the load testing execution using a centralized pacing mechanism.
func (e *Engine) Run(ctx context.Context) {
	var wg sync.WaitGroup
	
	// Create a context that respects the duration
	runCtx, cancel := context.WithTimeout(ctx, e.duration)
	defer cancel()

	// ticks channel will signal workers to perform an attack
	ticks := make(chan struct{})

	// Start workers
	// Each worker waits for a tick and performs an attack
	for i := 0; i < e.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-runCtx.Done():
					return
				case <-ticks:
					res := e.attacker.Attack(runCtx)
					e.stats.Add(res)
				}
			}
		}()
	}

	// Pacer: generates ticks at the target QPS
	go func() {
		defer close(ticks)
		
		// If QPS is 0, we shouldn't send any ticks or maybe max speed?
		// For now, assume QPS > 0 as per CLI defaults.
		if e.qps <= 0 {
			return
		}

		interval := time.Second / time.Duration(e.qps)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				select {
				case ticks <- struct{}{}:
				default:
					// If we can't send a tick, all workers are busy.
					// This indicates the system under test is slow or 
					// we don't have enough workers to maintain the rate.
				}
			}
		}
	}()

	// Wait for all workers to finish (triggered by duration timeout and channel close)
	wg.Wait()
}
