package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type DBManager struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewDBManager(
	logger *zap.Logger,
	db *sql.DB,
	dsn string,
	migrationsPath string,
) (*DBManager, error) {
	m := &DBManager{
		logger: logger,
		db:     db,
	}

	if err := m.applyMigrations(dsn, migrationsPath); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	return m, nil
}

func (m *DBManager) Close() error {
	return m.db.Close()
}

func (m *DBManager) applyMigrations(dsn string, migrationsPath string) error {
	mg, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("failed to initialize database for migrations: %w", err)
	}
	err = mg.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		m.logger.Info("No new migrations to apply")
	} else if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

func (m *DBManager) GetByOriginalURL(ctx context.Context, origURL string) (string, error) {
	q := "SELECT short_id FROM url_storage WHERE original_url = $1"
	return m.getByQuery(ctx, q, origURL)
}

func (m *DBManager) GetByShortID(ctx context.Context, shortID string) (string, error) {
	q := "SELECT original_url FROM url_storage WHERE short_id = $1"
	return m.getByQuery(ctx, q, shortID)
}

func (m *DBManager) getByQuery(ctx context.Context, q string, args ...any) (string, error) {
	row := m.db.QueryRowContext(ctx, q, args...)

	var url string
	err := row.Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrDataNotFoundInDB
	} else if err != nil {
		return "", fmt.Errorf("failed to scan query result row: %w", err)
	}
	return url, nil
}

func (m *DBManager) Persist(ctx context.Context, origURL, shortID string) error {
	q := "INSERT INTO url_storage (original_url, short_id) VALUES ($1, $2)"
	_, err := m.db.ExecContext(ctx, q, origURL, shortID)
	if err != nil {
		return fmt.Errorf("failed to persist binding (%s, %s) to db: %w", origURL, shortID, err)
	}
	return nil
}

func (m *DBManager) PersistBatch(ctx context.Context, binds *[]URLBind) error {
	insertSQL := "INSERT INTO url_storage (original_url, short_id) VALUES ($1, $2)"

	trx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error on begin transaction: %w", err)
	}
	defer func() {
		if err := trx.Rollback(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				m.logger.Error("failed to rollback transaction", zap.Error(err))
			}
		}
	}()

	stmt, err := trx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return fmt.Errorf("error on prepare statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				m.logger.Error("failed to close statement", zap.Error(err))
			}
		}
	}()

	for _, b := range *binds {
		if _, eErr := stmt.ExecContext(ctx, b.OrigURL, b.ShortID); eErr != nil {
			return fmt.Errorf("failed to persist batch record `%v` to db: %w", b, eErr)
		}
	}

	if cErr := trx.Commit(); cErr != nil {
		return fmt.Errorf("error on commiting transaction: %w", cErr)
	}
	return nil
}

func (m *DBManager) Ping(ctx context.Context) error {
	if err := m.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}
	return nil
}

var (
	ErrDataNotFoundInDB = errors.New("data not found in db")
)
