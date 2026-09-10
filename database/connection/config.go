// Package connection provides PostgreSQL connection management.
package connection

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config defines all configuration parameters for establishing and managing
// PostgreSQL database connections.
type Config struct {
	// Host is the database server hostname or IP address.
	Host string

	// Port is the database server port number.
	Port int

	// Database is the name of the database to connect to.
	Database string

	// Username is the database user for authentication.
	Username string

	// Password is the database user's password.
	Password string

	// SSLMode controls SSL/TLS behavior.
	// Valid values: "disable", "require", "verify-ca", "verify-full".
	SSLMode string

	// MaxConnections defines the maximum size of the connection pool.
	MaxConnections int32

	// MinConnections defines the minimum number of connections to maintain.
	MinConnections int32

	// MaxConnectionLifetime is the maximum amount of time a connection may be reused.
	MaxConnectionLifetime time.Duration

	// MaxConnectionIdleTime is the maximum amount of time a connection may be idle.
	MaxConnectionIdleTime time.Duration

	// ConnectTimeout is the maximum time to wait for a connection attempt.
	ConnectTimeout time.Duration

	// StatementTimeout is the default timeout for SQL statements.
	StatementTimeout time.Duration

	// IdleInTransactionSessionTimeout is the maximum allowed idle time
	// for a session within an open transaction.
	IdleInTransactionSessionTimeout time.Duration

	// HealthCheckPeriod determines how often the pool's internal health checker runs.
	HealthCheckPeriod time.Duration
}

// ConfigFromEnv builds a Config from standard DB_* environment variables.
// Fields not set in the environment retain their zero values; call Validate
// afterwards to apply defaults.
func ConfigFromEnv() *Config {
	c := &Config{
		Host:     os.Getenv("DB_HOST"),
		Database: os.Getenv("DB_NAME"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
	}
	if p := os.Getenv("DB_PORT"); p != "" {
		if port, err := strconv.Atoi(p); err == nil {
			c.Port = port
		}
	}
	return c
}

// Validate checks the configuration for required fields and fills in defaults.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Port == 0 {
		c.Port = 5432
	}
	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Username == "" {
		return fmt.Errorf("database username is required")
	}

	switch c.SSLMode {
	case "disable", "require", "verify-ca", "verify-full":
		// valid
	case "":
		c.SSLMode = "require"
	default:
		return fmt.Errorf("invalid SSLMode %q: must be disable, require, verify-ca, or verify-full", c.SSLMode)
	}

	if c.MaxConnections <= 0 {
		c.MaxConnections = 50
	}
	if c.MinConnections < 0 {
		c.MinConnections = 0
	}
	if c.MinConnections > c.MaxConnections {
		c.MinConnections = c.MaxConnections
	}
	if c.MaxConnectionLifetime <= 0 {
		c.MaxConnectionLifetime = time.Hour
	}
	if c.MaxConnectionIdleTime <= 0 {
		c.MaxConnectionIdleTime = 30 * time.Minute
	}
	if c.ConnectTimeout <= 0 {
		c.ConnectTimeout = 10 * time.Second
	}
	if c.StatementTimeout <= 0 {
		c.StatementTimeout = 30 * time.Second
	}
	if c.IdleInTransactionSessionTimeout <= 0 {
		c.IdleInTransactionSessionTimeout = 5 * time.Second
	}
	if c.HealthCheckPeriod <= 0 {
		c.HealthCheckPeriod = 30 * time.Second
	}
	return nil
}

// DSN constructs a PostgreSQL connection string.
//
// WARNING: the returned string contains the plaintext password. Do not log it.
// Prefer passing this only to pgxpool.ParseConfig and discarding it immediately.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
		int(c.ConnectTimeout.Seconds()),
	)
}

// SafeString returns a loggable representation that omits the password.
func (c *Config) SafeString() string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Database, c.SSLMode)
}
