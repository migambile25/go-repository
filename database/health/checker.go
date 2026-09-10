// Package health provides comprehensive health checking for database connectivity.
package health

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Status represents the health status of a database connection.
type Status int32

const (
	StatusUnknown   Status = 0
	StatusHealthy   Status = 1
	StatusUnhealthy Status = 2
)

func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// Checker performs health checks on a database connection pool.
// All methods are safe for concurrent use.
type Checker struct {
	pool      *pgxpool.Pool
	status    atomic.Int32 // stores Status values
	lastCheck atomic.Int64 // stores UnixNano of last check
	interval  time.Duration
}

// NewChecker creates a new health checker.
func NewChecker(pool *pgxpool.Pool, interval time.Duration) *Checker {
	c := &Checker{
		pool:     pool,
		interval: interval,
	}
	c.status.Store(int32(StatusUnknown))
	return c
}

// Check performs a single health check by pinging the database.
// The provided ctx controls the ping timeout.
func (c *Checker) Check(ctx context.Context) error {
	err := c.pool.Ping(ctx)
	c.lastCheck.Store(time.Now().UnixNano())

	if err != nil {
		c.status.Store(int32(StatusUnhealthy))
		return fmt.Errorf("health check failed: %w", err)
	}

	c.status.Store(int32(StatusHealthy))
	return nil
}

// Status returns the current health status.
func (c *Checker) Status() Status {
	return Status(c.status.Load())
}

// IsHealthy reports whether the last check succeeded.
func (c *Checker) IsHealthy() bool {
	return c.Status() == StatusHealthy
}

// LastCheckTime returns the timestamp of the most recent health check,
// or the zero value if no check has been performed yet.
func (c *Checker) LastCheckTime() time.Time {
	ns := c.lastCheck.Load()
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

// StartBackground launches a goroutine that performs health checks on the
// configured interval. The goroutine exits when ctx is cancelled.
//
// Each individual ping is given its own timeout derived from the interval
// so a slow server does not block the next scheduled check indefinitely.
func (c *Checker) StartBackground(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Give each ping at most half the interval to complete.
				pingCtx, cancel := context.WithTimeout(ctx, c.interval/2)
				_ = c.Check(pingCtx)
				cancel()
			}
		}
	}()
}
