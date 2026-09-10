package transaction

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is satisfied by both *pgxpool.Pool and pgx.Tx.
// Repositories store this as their executor so they work
// identically inside and outside a transaction.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Connection returns the active transaction from ctx if one was injected by
// WithinTransaction or WithRetry, otherwise returns the pool.
//
// Usage in any repository method:
//
//	func (r *Repository) Create(ctx context.Context, u *User) error {
//	    _, err := transaction.Conn(ctx, r.pool).Exec(ctx, query, args...)
//	    return err
//	}
func Connection(ctx context.Context, pool *pgxpool.Pool) DBTX {
	if tx := ExtractTx(ctx); tx != nil {
		return tx
	}
	return pool
}