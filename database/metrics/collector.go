// Package metrics provides Prometheus metrics for database operations.
package metrics

import (
	"github.com/IBM/pgxpoolprometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector aggregates all database-related Prometheus metrics.
type Collector struct {
	poolCollector       prometheus.Collector
	queryDuration       *prometheus.HistogramVec
	transactionDuration *prometheus.HistogramVec
	errorsTotal         *prometheus.CounterVec
}

// NewCollector creates a new metrics Collector and registers it with reg.
// Pass prometheus.DefaultRegisterer for the global registry, or a custom
// registry for test isolation or multi-tenant setups.
//
// labels are additional static labels attached to the pool metrics
// (e.g. {"service": "myapp", "region": "us-east-1"}).
func NewCollector(pool *pgxpool.Pool, labels map[string]string, reg prometheus.Registerer) (*Collector, error) {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	c := &Collector{
		poolCollector: pgxpoolprometheus.NewCollector(pool, labels),
		queryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Duration of database queries in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"query_type", "table"},
		),
		transactionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_transaction_duration_seconds",
				Help:    "Duration of database transactions in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		errorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_errors_total",
				Help: "Total number of database errors by type.",
			},
			[]string{"error_type", "query_type"},
		),
	}

	for _, col := range []prometheus.Collector{
		c.poolCollector,
		c.queryDuration,
		c.transactionDuration,
		c.errorsTotal,
	} {
		if err := reg.Register(col); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// RecordQueryDuration records the duration of a query execution.
func (c *Collector) RecordQueryDuration(queryType, table string, durationSeconds float64) {
	c.queryDuration.WithLabelValues(queryType, table).Observe(durationSeconds)
}

// RecordTransactionDuration records the total duration of a transaction.
func (c *Collector) RecordTransactionDuration(operation string, durationSeconds float64) {
	c.transactionDuration.WithLabelValues(operation).Observe(durationSeconds)
}

// RecordError increments the error counter for the given error and query types.
func (c *Collector) RecordError(errorType, queryType string) {
	c.errorsTotal.WithLabelValues(errorType, queryType).Inc()
}
