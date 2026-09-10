// Package errors defines typed sentinel errors for database operations.
package errors

import "errors"

// Sentinel errors for database operations.
// Use errors.Is to test for these in calling code.
var (
	ErrConnectionFailed              = errors.New("database connection failed")
	ErrTransactionFailed             = errors.New("transaction failed")
	ErrQueryFailed                   = errors.New("query execution failed")
	ErrPoolExhausted                 = errors.New("connection pool exhausted")
	ErrReplicaUnavailable            = errors.New("read replica unavailable")
	ErrInvalidConfig                 = errors.New("invalid database configuration")
	ErrMigrationFailed               = errors.New("migration failed")
	ErrStatementTimeout              = errors.New("statement execution timeout")
	ErrConnectionClosed              = errors.New("connection closed")
	ErrTransactionAlreadyActive      = errors.New("transaction already active in context")
)

// IsFatalConnectionError reports whether err represents an unrecoverable
// connection problem that should not be retried without operator intervention.
func IsFatalConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionClosed) ||
		errors.Is(err, ErrInvalidConfig)
}
