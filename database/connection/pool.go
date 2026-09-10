package connection

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

// Pool wraps pgxpool.Pool with enhanced monitoring capabilities.
type Pool struct {
	*pgxpool.Pool
	config *Config
}

// PoolMetrics provides a point-in-time snapshot of connection pool statistics.
// Every call to Statistics() returns a freshly allocated value — it is safe to
// hold references to multiple snapshots concurrently.
type PoolMetrics struct {
	TotalConns              int32
	IdleConns               int32
	AcquiredConns           int32
	ConstructingConns       int32
	MaxConns                int32
	AcquireCount            int64
	CanceledAcquireCount    int64
	EmptyAcquireCount       int64
	MaxLifetimeDestroyCount int64
	MaxIdleDestroyCount     int64
}

// NewPool creates and initializes a new connection pool with the given configuration.
func NewPool(ctx context.Context, cfg *Config) (*Pool, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Parse using the DSN. The DSN is only passed to pgx internals and is not
	// retained after this call.
	pgxConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pgxConfig.MaxConns = cfg.MaxConnections
	pgxConfig.MinConns = cfg.MinConnections
	pgxConfig.MaxConnLifetime = cfg.MaxConnectionLifetime
	pgxConfig.MaxConnIdleTime = cfg.MaxConnectionIdleTime
	pgxConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Configure session parameters on every new connection.
	pgxConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Exec(ctx, "SET TIME ZONE 'UTC'"); err != nil {
			return fmt.Errorf("failed to set time zone: %w", err)
		}
		if _, err := conn.Exec(ctx,
			fmt.Sprintf("SET statement_timeout = '%dms'", cfg.StatementTimeout.Milliseconds()),
		); err != nil {
			return fmt.Errorf("failed to set statement_timeout: %w", err)
		}
		if _, err := conn.Exec(ctx,
			fmt.Sprintf("SET idle_in_transaction_session_timeout = '%dms'",
				cfg.IdleInTransactionSessionTimeout.Milliseconds()),
		); err != nil {
			return fmt.Errorf("failed to set idle_in_transaction_session_timeout: %w", err)
		}
		return nil
	}

	// Attach a no-op tracer that silences pgx's internal query logging by default.
	// Replace this with a real tracelog.TraceLog + logger to enable SQL tracing.
	pgxConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   tracelog.LoggerFunc(func(_ context.Context, _ tracelog.LogLevel, _ string, _ map[string]any) {}),
		LogLevel: tracelog.LogLevelNone,
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Eagerly verify connectivity so callers get a clear error at startup.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{Pool: pool, config: cfg}, nil
}

// Statistics returns a fresh, concurrency-safe snapshot of pool metrics.
func (p *Pool) Statistics() *PoolMetrics {
	s := p.Stat()
	return &PoolMetrics{
		TotalConns:              s.TotalConns(),
		IdleConns:               s.IdleConns(),
		AcquiredConns:           s.AcquiredConns(),
		ConstructingConns:       s.ConstructingConns(),
		MaxConns:                s.MaxConns(),
		AcquireCount:            s.AcquireCount(),
		CanceledAcquireCount:    s.CanceledAcquireCount(),
		EmptyAcquireCount:       s.EmptyAcquireCount(),
		MaxLifetimeDestroyCount: s.MaxLifetimeDestroyCount(),
		MaxIdleDestroyCount:     s.MaxIdleDestroyCount(),
	}
}

// Config returns the configuration used to create this pool (read-only).
func (p *Pool) Config() *Config { return p.config }
