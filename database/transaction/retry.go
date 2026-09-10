package transaction

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsRetryableError reports whether err should trigger a transaction retry.
//
// Retryable SQLSTATE codes:
//   - 40001  serialization_failure
//   - 40P01  deadlock_detected
//   - 55P03  lock_not_available
//   - 57014  query_canceled (statement timeout)
//   - 53300  too_many_connections
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40001", "40P01", "55P03", "57014", "53300":
			return true
		}
	}

	// In pgx v5, ErrDeadConn was removed. Closed/broken connections surface as
	// net.ErrClosed wrapped inside a pgconn error. Check for it explicitly.
	if errors.Is(err, net.ErrClosed) {
		return true
	}

	// Context deadline exceeded is typically not retryable within the same
	// context, but we allow callers to wrap it in a fresh context and retry.
	// We intentionally do NOT retry context.Canceled — that signals intent.
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}

// WithRetry executes fn with automatic retries on transient failures.
// It uses truncated exponential backoff with full jitter to avoid thundering herds.
func (m *PostgresTransactionManager) WithRetry(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	var lastErr error

	for attempt := 0; attempt <= m.opts.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := backoffDelay(attempt, m.opts.BaseRetryDelay, m.opts.MaxRetryDelay)
			select {
			case <-ctx.Done():
				return fmt.Errorf("context canceled during retry backoff: %w", ctx.Err())
			case <-time.After(delay):
			}
		}

		err := m.WithinTransaction(ctx, fn)
		if err == nil {
			return nil
		}
		if !IsRetryableError(err) {
			return err
		}
		lastErr = err
	}

	return fmt.Errorf("transaction failed after %d retries: %w", m.opts.MaxRetries, lastErr)
}

// backoffDelay returns a jittered delay for the given retry attempt.
// Strategy: full jitter — sleep a random duration in [0, min(cap, base * 2^attempt)).
func backoffDelay(attempt int, base, cap time.Duration) time.Duration {
	// Clamp the exponent to avoid overflow on large attempt counts.
	exp := attempt - 1
	if exp > 30 {
		exp = 30
	}
	ceiling := base * time.Duration(1<<uint(exp))
	if ceiling > cap || ceiling <= 0 {
		ceiling = cap
	}
	// rand.N is the idiomatic Go 1.22+ global PRNG function; no seeding required.
	return rand.N(ceiling)
}