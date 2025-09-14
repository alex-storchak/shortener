package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

type DBManager struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewDBManager(logger *zap.Logger, db *sql.DB) *DBManager {
	logger = logger.With(
		zap.String("component", "db manager"),
	)
	return &DBManager{
		logger: logger,
		db:     db,
	}
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
		return "", err
	}
	return url, nil
}

func (m *DBManager) Persist(ctx context.Context, origURL, shortID string) error {
	q := "INSERT INTO url_storage (original_url, short_id) VALUES ($1, $2)"
	_, err := m.db.ExecContext(ctx, q, origURL, shortID)
	if err != nil {
		m.logger.Error("Can't persist binding to db", zap.Error(err))
		return fmt.Errorf("error persisting binding to db: %w", err)
	}
	return nil
}

var (
	ErrDataNotFoundInDB = errors.New("data not found in db")
)
