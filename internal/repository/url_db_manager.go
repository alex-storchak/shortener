package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type URLDBManager struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewURLDBManager(logger *zap.Logger, db *sql.DB) *URLDBManager {
	return &URLDBManager{
		logger: logger,
		db:     db,
	}
}

func (m *URLDBManager) Close() error {
	return m.db.Close()
}

func (m *URLDBManager) GetByOriginalURL(ctx context.Context, origURL string) (string, error) {
	q := "SELECT short_id FROM url_storage WHERE original_url = $1"
	return m.getByQuery(ctx, q, origURL)
}

func (m *URLDBManager) GetByShortID(ctx context.Context, shortID string) (string, error) {
	q := "SELECT original_url FROM url_storage WHERE short_id = $1"
	return m.getByQuery(ctx, q, shortID)
}

func (m *URLDBManager) getByQuery(ctx context.Context, q string, args ...any) (string, error) {
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

func (m *URLDBManager) Persist(ctx context.Context, r *model.URLStorageRecord) error {
	q := "INSERT INTO url_storage (original_url, short_id, user_id) SELECT $1, $2, id FROM auth_user where user_uuid = $3"
	_, err := m.db.ExecContext(ctx, q, r.OrigURL, r.ShortID, r.UserUUID)
	if err != nil {
		return fmt.Errorf("failed to persist binding (%s, %s, %s) to db: %w", r.OrigURL, r.ShortID, r.UserUUID, err)
	}
	return nil
}

func (m *URLDBManager) PersistBatch(ctx context.Context, binds *[]model.URLStorageRecord) error {
	insertSQL := "INSERT INTO url_storage (original_url, short_id, user_id) SELECT $1, $2, id FROM auth_user where user_uuid = $3"

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
		if _, eErr := stmt.ExecContext(ctx, b.OrigURL, b.ShortID, b.UserUUID); eErr != nil {
			return fmt.Errorf("failed to persist batch record `%v` to db: %w", b, eErr)
		}
	}

	if cErr := trx.Commit(); cErr != nil {
		return fmt.Errorf("error on commiting transaction: %w", cErr)
	}
	return nil
}

func (m *URLDBManager) Ping(ctx context.Context) error {
	if err := m.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}
	return nil
}

func (m *URLDBManager) GetByUserUUID(ctx context.Context, userUUID string) (*[]model.URLStorageRecord, error) {
	urls := make([]model.URLStorageRecord, 0)

	q := "SELECT original_url, short_id, user_uuid FROM url_storage JOIN auth_user ON id = user_id WHERE user_uuid = $1"
	rows, err := m.db.QueryContext(ctx, q, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user urls from db: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var r model.URLStorageRecord
		err = rows.Scan(&r.OrigURL, &r.ShortID, &r.UserUUID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user url from db: %w", err)
		}
		urls = append(urls, r)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to get user urls from db: %w", err)
	}
	return &urls, nil
}

var (
	ErrDataNotFoundInDB = errors.New("data not found in db")
)
