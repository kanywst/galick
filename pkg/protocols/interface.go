// Package protocols defines the interfaces for different load testing protocols.
package protocols

import (
	"context"
	"github.com/kanywst/galick/pkg/metrics"
)

// Attacker is the interface that protocol implementations must satisfy.
type Attacker interface {
	// Attack performs a single request and returns the result.
	Attack(ctx context.Context) metrics.Result
	
	// Name returns the protocol name (e.g., "http", "grpc").
	Name() string
}