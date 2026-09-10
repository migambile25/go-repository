// Package transaction provides a robust transaction management system.
package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Manager defines the interface for transaction management.
type Manager interface {
	// WithinTransaction executes fn inside a single database transaction.
	// The transaction is committed if fn returns nil, or rolled back otherwise.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	// WithRetry wraps WithinTransaction with automatic retries on transient errors.
	WithRetry(ctx context.Context, fn func(ctx context.Context) error) error
}

// TransactionOptions configures transaction behavior.
type TransactionOptions struct {
	IsolationLevel pgx.TxIsoLevel
	AccessMode     pgx.TxAccessMode
	MaxRetries     int
	BaseRetryDelay time.Duration
	MaxRetryDelay  time.Duration
}

// DefaultTransactionOptions returns sensible production defaults.
func DefaultTransactionOptions() *TransactionOptions {
	return &TransactionOptions{
		IsolationLevel: pgx.ReadCommitted,
		AccessMode:     pgx.ReadWrite,
		MaxRetries:     3,
		BaseRetryDelay: 100 * time.Millisecond,
		MaxRetryDelay:  2 * time.Second,
	}
}

// PostgresTransactionManager implements Manager for PostgreSQL.
type PostgresTransactionManager struct {
	pool *pgxpool.Pool
	opts *TransactionOptions
}

// NewPostgresTransactionManager creates a new transaction manager.
// If opts is nil, DefaultTransactionOptions is used.
func NewPostgresTransactionManager(pool *pgxpool.Pool, opts *TransactionOptions) *PostgresTransactionManager {
	if opts == nil {
		opts = DefaultTransactionOptions()
	}
	return &PostgresTransactionManager{pool: pool, opts: opts}
}

// WithinTransaction executes fn within a database transaction.
//
// The implementation avoids the subtle named-return / defer interaction bug:
// instead of relying on a named error return variable being visible inside
// defer, we perform rollback/commit explicitly so the control flow is obvious
// and correct even when fn panics.
func (m *PostgresTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) (retErr error) {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   m.opts.IsolationLevel,
		AccessMode: m.opts.AccessMode,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure the transaction is always finalised.
	defer func() {
		if p := recover(); p != nil {
			// Roll back on panic, then re-panic so the caller can observe it.
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if retErr != nil {
			// Best-effort rollback; we prefer the original error over a rollback error.
			_ = tx.Rollback(ctx)
		}
	}()

	txCtx := context.WithValue(ctx, transactionKey{}, tx)

	if retErr = fn(txCtx); retErr != nil {
		return retErr // defer will rollback
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

type transactionKey struct{}

// ExtractTx retrieves the active transaction from the context, or nil if none.
func ExtractTx(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(transactionKey{}).(pgx.Tx)
	return tx
}
