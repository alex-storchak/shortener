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

func NewDB(cfg *config.DB, migrationsPath string, l *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := applyMigrations(cfg.DSN, migrationsPath, l); err != nil {
		return nil, err
	}

	return db, nil
}

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
