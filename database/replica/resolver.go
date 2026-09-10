// Package replica implements read replica load balancing and failover.
package replica

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Resolver manages multiple read replicas and provides intelligent routing.
// All methods are safe for concurrent use.
type Resolver struct {
	primary      *pgxpool.Pool
	replicas     []*pgxpool.Pool
	replicaIndex atomic.Uint64

	mu      sync.RWMutex
	healthy []bool // protected by mu
}

// NewResolver creates a new Resolver.
// All replicas are initially assumed healthy.
func NewResolver(primary *pgxpool.Pool, replicas ...*pgxpool.Pool) *Resolver {
	healthy := make([]bool, len(replicas))
	for i := range healthy {
		healthy[i] = true
	}
	return &Resolver{
		primary:  primary,
		replicas: replicas,
		healthy:  healthy,
	}
}

// Writer returns the primary pool for write operations.
func (r *Resolver) Writer() *pgxpool.Pool {
	return r.primary
}

// Reader returns a healthy replica using round-robin selection.
// Falls back to the primary if no replicas are configured or all are unhealthy.
//
// The round-robin implementation uses a pre-increment so index 0 is reachable:
// on the very first call Add(1) returns 1, and 1 % N == 0 when N == 1.
// For N > 1 the sequence cycles fairly starting at index (1 % N).
func (r *Resolver) Reader(_ context.Context) *pgxpool.Pool {
	if len(r.replicas) == 0 {
		return r.primary
	}

	n := uint64(len(r.replicas))
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Attempt each replica in round-robin order, skipping unhealthy ones.
	for range r.replicas {
		idx := r.replicaIndex.Add(1) % n
		if r.healthy[idx] {
			return r.replicas[idx]
		}
	}

	// All replicas unhealthy — fall back to primary.
	return r.primary
}

// PerformHealthChecks pings every replica and updates the health map atomically.
func (r *Resolver) PerformHealthChecks(ctx context.Context) {
	// Collect results outside the lock to avoid holding it during network I/O.
	results := make([]bool, len(r.replicas))
	for i, replica := range r.replicas {
		results[i] = replica.Ping(ctx) == nil
	}

	r.mu.Lock()
	copy(r.healthy, results)
	r.mu.Unlock()
}

// HealthStatus returns a snapshot of each replica's health keyed by pool index.
func (r *Resolver) HealthStatus() []bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]bool, len(r.healthy))
	copy(out, r.healthy)
	return out
}

// StartHealthCheckBackground starts periodic replica health checks.
// The goroutine exits when ctx is cancelled.
func (r *Resolver) StartHealthCheckBackground(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.PerformHealthChecks(ctx)
			}
		}
	}()
}
