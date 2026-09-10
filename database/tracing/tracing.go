// Package tracing provides OpenTelemetry integration for database operations.
package tracing

import (
	"context"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps OpenTelemetry tracing for database operations.
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new database tracer using the global OTel provider.
// serviceName is used as the instrumentation scope name prefix.
func NewTracer(serviceName string) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(serviceName + ".database"),
	}
}

// ConfigurePool attaches an OpenTelemetry tracer to the pgxpool configuration
// so that every query and batch operation is automatically instrumented.
// Call this before creating the pool with pgxpool.NewWithConfig.
func (t *Tracer) ConfigurePool(config *pgxpool.Config) {
	config.ConnConfig.Tracer = otelpgx.NewTracer()
}

// StartSpan begins a new span for a named database operation.
// The caller is responsible for ending the span:
//
//	ctx, span := tracer.StartSpan(ctx, "users.FindByID", query)
//	defer span.End()
func (t *Tracer) StartSpan(ctx context.Context, operation, query string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, operation,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", query),
		),
	)
}
