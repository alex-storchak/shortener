package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/db/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func NewDB(cfg *config.Config, migrationsPath string, zl *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := applyMigrations(cfg.DSN, migrationsPath, zl); err != nil {
		return nil, err
	}

	return db, nil
}

func applyMigrations(dsn string, migrationsPath string, zl *zap.Logger) error {
	mg, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("failed to initialize database for migrations: %w", err)
	}
	err = mg.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		zl.Info("No new migrations to apply")
	} else if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}
