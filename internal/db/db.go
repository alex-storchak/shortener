// Package db provides database connectivity and migration utilities for the URL shortener service.
// It handles PostgreSQL database connection management and schema migrations using the migrate tool.
//
// The package supports:
//   - Database connection pooling and configuration
//   - Automatic schema migrations on application startup
//   - PostgreSQL-specific optimizations using pgx driver
//   - Migration versioning and rollback capabilities
//
// Key Features:
//   - Connection management with proper error handling
//   - Idempotent migration application (safe for multiple runs)
//   - Structured logging for migration events
package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
)

// NewDB creates a new PostgreSQL database connection and applies pending migrations.
// It initializes the database connection pool using the pgx driver and ensures
// the database schema is up-to-date by running all pending migrations.
//
// Parameters:
//   - cfg: Database configuration containing connection DSN and settings
//   - l: Structured logger for logging operations
//
// Returns:
//   - *sql.DB: Configured database connection pool
//   - error: nil on success, or error if connection or migrations fail
func NewDB(cfg *config.DB, l *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := applyMigrations(cfg.DSN, cfg.MigrationsPath, l); err != nil {
		return nil, err
	}

	return db, nil
}

// applyMigrations applies all pending database migrations
// to bring the schema to the latest version.
// It uses the golang-migrate library to handle migration execution and version tracking.
//
// Parameters:
//   - dsn: Data Source Name for database connection
//   - migrationsPath: File system path to migration files (e.g., "file://migrations")
//   - l: Structured logger for logging operations
//
// Returns:
//   - error: nil on success, migrate.ErrNoChange if no migrations needed, or error if migrations fail
//
// Behavior:
//   - Logs informational message when no migrations are needed
//   - Applies migrations in transactional manner where supported
//   - Handles forward (up) migrations only
func applyMigrations(dsn string, migrationsPath string, l *zap.Logger) error {
	mg, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("init db for migrations: %w", err)
	}
	err = mg.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		l.Info("No new migrations to apply")
	} else if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
