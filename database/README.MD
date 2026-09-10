# database

A production-grade PostgreSQL package for Go built on [`pgx v5`](https://github.com/jackc/pgx).
No ORM, no reflection in hot paths — just a thin, well-tested layer over pgxpool with
connection pooling, transaction management, read-replica routing, migrations,
Prometheus metrics, and OpenTelemetry tracing.

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Connection Pool](#connection-pool)
- [Transaction Management](#transaction-management)
- [Read Replicas](#read-replicas)
- [Migrations](#migrations)
- [Health Checks](#health-checks)
- [Metrics (Prometheus)](#metrics-prometheus)
- [Tracing (OpenTelemetry)](#tracing-opentelemetry)
- [Error Handling](#error-handling)
- [Running Tests](#running-tests)
- [Project Structure](#project-structure)
- [License](#license)

---

## Features

| Feature | Details |
|---|---|
| **Connection pooling** | pgxpool with configurable min/max, lifetime, idle, and health-check periods |
| **Transaction manager** | Auto-commit/rollback, panic recovery, context propagation |
| **Retry framework** | Exponential back-off with full jitter; retries serialization failures, deadlocks, and timeouts |
| **Read replicas** | Round-robin load balancing with background health checks and automatic primary fallback |
| **Migrations** | golang-migrate v4 with file source; `Up`, `Down`, `Steps`, and `Version` |
| **Health checks** | Liveness and readiness probes with per-ping timeouts |
| **Prometheus metrics** | Pool stats via pgxpoolprometheus; query/transaction duration histograms; error counters |
| **OpenTelemetry tracing** | otelpgx pool instrumentation and manual span helpers |
| **Typed errors** | Sentinel errors for clean `errors.Is` usage |
| **Race-safe** | All shared state protected; passes `go test -race` |

---

## Requirements

- Go 1.22 or later
- PostgreSQL 13 or later

---

## Installation

```bash
go get github.com/hekimapro/database
```

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/hekimapro/database/connection"
    "github.com/hekimapro/database/transaction"
)

func main() {
    ctx := context.Background()

    pool, err := connection.NewPool(ctx, &connection.Config{
        Host:     "localhost",
        Database: "myapp",
        Username: "postgres",
        Password: "password",
        SSLMode:  "disable", // use "require" in production
    })
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    tm := transaction.NewPostgresTransactionManager(pool.Pool, nil)

    err = tm.WithRetry(ctx, func(txCtx context.Context) error {
        _, err := pool.Exec(txCtx, "INSERT INTO users (name) VALUES ($1)", "Alice")
        return err
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Done.")
}
```

---

## Configuration

### Struct fields

```go
cfg := &connection.Config{
    Host:                            "localhost",
    Port:                            5432,           // default: 5432
    Database:                        "myapp",
    Username:                        "postgres",
    Password:                        "password",
    SSLMode:                         "require",      // disable | require | verify-ca | verify-full
    MaxConnections:                  50,             // default: 50
    MinConnections:                  5,              // default: 0
    MaxConnectionLifetime:           time.Hour,      // default: 1h
    MaxConnectionIdleTime:           30*time.Minute, // default: 30m
    ConnectTimeout:                  10*time.Second, // default: 10s
    StatementTimeout:                30*time.Second, // default: 30s
    IdleInTransactionSessionTimeout: 5*time.Second,  // default: 5s
    HealthCheckPeriod:               30*time.Second, // default: 30s
}
```

### Environment variables

`ConfigFromEnv()` reads standard `DB_*` variables:

| Variable | Field |
|---|---|
| `DB_HOST` | Host |
| `DB_PORT` | Port |
| `DB_NAME` | Database |
| `DB_USER` | Username |
| `DB_PASSWORD` | Password |
| `DB_SSL_MODE` | SSLMode |

```go
cfg := connection.ConfigFromEnv()
pool, err := connection.NewPool(ctx, cfg)
```

### Logging the config safely

```go
log.Printf("connecting: %s", cfg.SafeString()) // password is never included
```

---

## Connection Pool

```go
pool, err := connection.NewPool(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

// Point-in-time snapshot — safe to hold across goroutines
stats := pool.Statistics()
fmt.Printf("idle=%d acquired=%d total=%d\n",
    stats.IdleConns, stats.AcquiredConns, stats.TotalConns)
```

`pool.Pool` exposes the full `*pgxpool.Pool` API (`Exec`, `Query`, `QueryRow`,
`Acquire`, `BeginTx`, etc.).

Every connection is automatically configured with:
- `TIME ZONE = 'UTC'`
- `statement_timeout` from `Config.StatementTimeout`
- `idle_in_transaction_session_timeout` from `Config.IdleInTransactionSessionTimeout`

---

## Transaction Management

### Basic transaction

```go
tm := transaction.NewPostgresTransactionManager(pool.Pool, nil)

err := tm.WithinTransaction(ctx, func(txCtx context.Context) error {
    _, err := pool.Exec(txCtx, "UPDATE accounts SET balance = balance - 100 WHERE id = $1", 1)
    if err != nil {
        return err // triggers automatic rollback
    }
    _, err = pool.Exec(txCtx, "UPDATE accounts SET balance = balance + 100 WHERE id = $1", 2)
    return err // commit on nil, rollback on error
})
```

### With retries

```go
err := tm.WithRetry(ctx, func(txCtx context.Context) error {
    // automatically retried on: serialization_failure, deadlock, lock_not_available,
    // statement timeout, too_many_connections, ErrDeadConn, DeadlineExceeded
    return doWork(txCtx)
})
```

### Custom options

```go
opts := &transaction.TransactionOptions{
    IsolationLevel: pgx.Serializable,
    AccessMode:     pgx.ReadWrite,
    MaxRetries:     5,
    BaseRetryDelay: 50 * time.Millisecond,
    MaxRetryDelay:  3 * time.Second,
}
tm := transaction.NewPostgresTransactionManager(pool.Pool, opts)
```

### Accessing the transaction from a helper

```go
func insertUser(ctx context.Context, pool *pgxpool.Pool, name string) error {
    // Uses the active transaction if present, otherwise falls back to pool.
    if tx := transaction.ExtractTx(ctx); tx != nil {
        _, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", name)
        return err
    }
    _, err := pool.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", name)
    return err
}
```

---

## Read Replicas

```go
resolver := replica.NewResolver(primaryPool, replica1Pool, replica2Pool)

// Start background health checks every 15 seconds.
resolver.StartHealthCheckBackground(ctx, 15*time.Second)

// Writes go to primary.
resolver.Writer().Exec(ctx, "INSERT INTO events (name) VALUES ($1)", "click")

// Reads are round-robin across healthy replicas, with automatic fallback to primary.
reader := resolver.Reader(ctx)
reader.QueryRow(ctx, "SELECT COUNT(*) FROM events").Scan(&n)

// Inspect health.
for i, healthy := range resolver.HealthStatus() {
    fmt.Printf("replica %d: healthy=%v\n", i, healthy)
}
```

If all replicas are unhealthy, `Reader` returns the primary automatically.

---

## Migrations

```go
m, err := migration.NewMigrator(
    "file:///app/migrations",
    "postgres://postgres:password@localhost:5432/myapp?sslmode=disable",
)
if err != nil {
    log.Fatal(err)
}
defer m.Close()

// Apply all pending migrations.
if err := m.Up(); err != nil {
    log.Fatal(err)
}

// Check current version.
version, dirty, err := m.Version()
fmt.Printf("version=%d dirty=%v err=%v\n", version, dirty, err)

// Rollback one step.
if err := m.Down(); err != nil {
    log.Fatal(err)
}
```

Migration files follow the `golang-migrate` naming convention:

```
migrations/
  000001_create_users.up.sql
  000001_create_users.down.sql
  000002_add_email.up.sql
  000002_add_email.down.sql
```

---

## Health Checks

```go
checker := health.NewChecker(pool.Pool, 10*time.Second)

// Start background checks; exits when ctx is cancelled.
checker.StartBackground(ctx)

// HTTP handlers
http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
    if checker.IsHealthy() {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintln(w, "ok")
        return
    }
    w.WriteHeader(http.StatusServiceUnavailable)
    fmt.Fprintf(w, "unhealthy since %s\n", checker.LastCheckTime().Format(time.RFC3339))
})

// One-shot check.
if err := checker.Check(ctx); err != nil {
    log.Printf("db unhealthy: %v", err)
}

fmt.Println(checker.Status()) // "healthy" | "unhealthy" | "unknown"
```

---

## Metrics (Prometheus)

```go
import (
    "github.com/hekimapro/database/metrics"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "net/http"
)

collector, err := metrics.NewCollector(
    pool.Pool,
    map[string]string{"service": "myapp"},
    prometheus.DefaultRegisterer,
)
if err != nil {
    log.Fatal(err)
}

// Record durations from your repository layer.
collector.RecordQueryDuration("SELECT", "users", elapsed.Seconds())
collector.RecordTransactionDuration("transfer", elapsed.Seconds())
collector.RecordError("serialization_failure", "UPDATE")

// Expose the /metrics endpoint.
http.Handle("/metrics", promhttp.Handler())
```

### Exported metrics

| Metric | Type | Labels |
|---|---|---|
| `database_query_duration_seconds` | Histogram | `query_type`, `table` |
| `database_transaction_duration_seconds` | Histogram | `operation` |
| `database_errors_total` | Counter | `error_type`, `query_type` |
| pgxpool built-ins (via pgxpoolprometheus) | Gauges/Counters | custom labels |

---

## Tracing (OpenTelemetry)

```go
// Configure your OTel provider and exporter before calling this.
dbTracer := tracing.NewTracer("myapp")

// Instrument a pool (before pgxpool.NewWithConfig).
dbTracer.ConfigurePool(pgxConfig)

// Manual span for a named operation.
ctx, span := dbTracer.StartSpan(ctx, "users.FindByID",
    "SELECT id, name FROM users WHERE id = $1")
defer span.End()

row := pool.QueryRow(ctx, "SELECT id, name FROM users WHERE id = $1", userID)
```

Every query executed through the instrumented pool is automatically traced by
otelpgx with `db.system=postgresql` and the full query text.

---

## Error Handling

```go
import dberrors "github.com/hekimapro/database/errors"

if errors.Is(err, dberrors.ErrConnectionFailed) { ... }
if errors.Is(err, dberrors.ErrTransactionFailed) { ... }
if errors.Is(err, dberrors.ErrPoolExhausted)     { ... }
if errors.Is(err, dberrors.ErrMigrationFailed)   { ... }

if dberrors.IsFatalConnectionError(err) {
    // Do not retry; alert on-call.
}
```

For retryability at the pgx level, use `transaction.IsRetryableError`:

```go
if transaction.IsRetryableError(err) {
    // Safe to retry this error class.
}
```

---

## Running Tests

### Unit tests (no database required)

```bash
go test -race -cover ./...
```

### Integration tests (requires PostgreSQL)

```bash
export DB_HOST=localhost DB_NAME=testdb DB_USER=postgres DB_PASSWORD=password DB_SSL_MODE=disable
go test -tags integration -race -cover ./test/integration/...
```

### Benchmarks

```bash
go test -tags integration -bench=. -benchmem ./test/benchmark/...
```

---

## Project Structure

```
github.com/hekimapro/database/
├── README.md
├── go.mod
├── connection/
│   ├── config.go          — Config, ConfigFromEnv, Validate, DSN, SafeString
│   ├── pool.go            — Pool, NewPool, Statistics
│   └── pool_test.go
├── transaction/
│   ├── manager.go         — Manager interface, PostgresTransactionManager, ExtractTx
│   ├── manager_test.go
│   └── retry.go           — IsRetryableError, WithRetry, backoffDelay
├── migration/
│   └── migrator.go        — Migrator, Up, Down, Steps, Version, Close
├── metrics/
│   ├── collector.go       — Collector, NewCollector, Record*
│   └── collector_test.go
├── tracing/
│   └── tracing.go         — Tracer, NewTracer, ConfigurePool, StartSpan
├── health/
│   ├── checker.go         — Checker, Status, IsHealthy, StartBackground
│   └── checker_test.go
├── replica/
│   ├── resolver.go        — Resolver, Reader, Writer, HealthStatus, StartHealthCheckBackground
│   └── resolver_test.go
├── errors/
│   └── errors.go          — Sentinel errors, IsFatalConnectionError
├── examples/
│   ├── basic/main.go
│   ├── transaction/main.go
│   ├── replica/main.go
│   └── observability/main.go
└── test/
    ├── integration/integration_test.go
    └── benchmark/benchmark_test.go
```

---

## License

MIT — see [LICENSE](LICENSE) for details.
