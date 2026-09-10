// Package migration provides database schema migration capabilities.
package migration

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrator handles database schema migrations via golang-migrate.
type Migrator struct {
	migrate *migrate.Migrate
}

// NewMigrator creates a new Migrator.
//
//	sourceURL   — e.g. "file:///path/to/migrations"
//	databaseURL — e.g. "postgres://user:pass@localhost:5432/dbname?sslmode=require"
func NewMigrator(sourceURL, databaseURL string) (*Migrator, error) {
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}
	return &Migrator{migrate: m}, nil
}

// Up applies all pending migrations.
func (m *Migrator) Up() error {
	if err := m.migrate.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply up migrations: %w", err)
	}
	return nil
}

// Down reverts the most recently applied migration.
func (m *Migrator) Down() error {
	if err := m.migrate.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to revert migration: %w", err)
	}
	return nil
}

// Steps applies n migrations (positive = up, negative = down).
func (m *Migrator) Steps(n int) error {
	if err := m.migrate.Steps(n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply %d migration steps: %w", n, err)
	}
	return nil
}

// Version returns the current migration version and whether the database is
// currently in a dirty state.
func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

// Close releases resources held by the migrator.
// Both the source and database errors are reported; neither is silently dropped.
func (m *Migrator) Close() error {
	srcErr, dbErr := m.migrate.Close()
	return errors.Join(srcErr, dbErr)
}
