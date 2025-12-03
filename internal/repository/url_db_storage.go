package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

var (
	ErrDataNotFoundInDB = errors.New("data not found in db")
)

type DBURLStorage struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewDBURLStorage(logger *zap.Logger, db *sql.DB) *DBURLStorage {
	return &DBURLStorage{
		logger: logger,
		db:     db,
	}
}

func (s *DBURLStorage) Close() error {
	return s.db.Close()
}

func (s *DBURLStorage) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	return nil
}

func (s *DBURLStorage) getByOriginalURL(ctx context.Context, origURL string) (*model.URLStorageRecord, error) {
	q := `
		SELECT us.original_url, us.short_id, au.user_uuid, us.is_deleted 
		FROM url_storage us
		JOIN auth_user au ON au.id = us.user_id 
		WHERE us.original_url = $1
		AND us.is_deleted = FALSE
	`
	return s.getByQuery(ctx, q, origURL)
}

func (s *DBURLStorage) getByShortID(ctx context.Context, shortID string) (*model.URLStorageRecord, error) {
	q := `
		SELECT us.original_url, us.short_id, au.user_uuid, us.is_deleted 
		FROM url_storage us
		JOIN auth_user au ON au.id = us.user_id 
		WHERE us.short_id = $1
	`
	return s.getByQuery(ctx, q, shortID)
}

func (s *DBURLStorage) getByQuery(ctx context.Context, q string, args ...any) (*model.URLStorageRecord, error) {
	row := s.db.QueryRowContext(ctx, q, args...)

	var r model.URLStorageRecord
	err := row.Scan(&r.OrigURL, &r.ShortID, &r.UserUUID, &r.IsDeleted)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDataNotFoundInDB
	} else if err != nil {
		return nil, fmt.Errorf("scan query result row: %w", err)
	}
	return &r, nil
}

func (s *DBURLStorage) Get(ctx context.Context, url, searchByType string) (*model.URLStorageRecord, error) {
	var (
		r   *model.URLStorageRecord
		err error
	)
	if searchByType == OrigURLType {
		r, err = s.getByOriginalURL(ctx, url)
	} else if searchByType == ShortURLType {
		r, err = s.getByShortID(ctx, url)
		if err == nil && r.IsDeleted {
			return nil, ErrDataDeleted
		}
	}
	if errors.Is(err, ErrDataNotFoundInDB) {
		return nil, NewDataNotFoundError(ErrDataNotFoundInDB)
	} else if err != nil {
		return nil, fmt.Errorf("retrieve bind by url `%s` from db: %w", url, err)
	}
	return r, nil
}

func (s *DBURLStorage) Set(ctx context.Context, origURL, shortID, userUUID string) error {
	q := `
		INSERT INTO url_storage (original_url, short_id, user_id) 
		SELECT $1, $2, id 
		FROM auth_user 
		WHERE user_uuid = $3
	`
	_, err := s.db.ExecContext(ctx, q, origURL, shortID, userUUID)
	if err != nil {
		return fmt.Errorf("persist binding (%s, %s, %s) to db: %w", origURL, shortID, userUUID, err)
	}
	return nil
}

func (s *DBURLStorage) BatchSet(ctx context.Context, records []model.URLStorageRecord) error {
	insertSQL := `
		INSERT INTO url_storage (original_url, short_id, user_id) 
		SELECT $1, $2, id 
		FROM auth_user 
		WHERE user_uuid = $3
	`
	trx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := trx.Rollback(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				s.logger.Error("failed to rollback transaction", zap.Error(err))
			}
		}
	}()

	stmt, err := trx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				s.logger.Error("failed to close statement", zap.Error(err))
			}
		}
	}()

	for _, b := range records {
		if _, eErr := stmt.ExecContext(ctx, b.OrigURL, b.ShortID, b.UserUUID); eErr != nil {
			return fmt.Errorf("persist batch record `%v` to db: %w", b, eErr)
		}
	}

	if cErr := trx.Commit(); cErr != nil {
		return fmt.Errorf("commiting transaction: %w", cErr)
	}
	return nil
}

func (s *DBURLStorage) GetByUserUUID(ctx context.Context, userUUID string) ([]*model.URLStorageRecord, error) {
	urls := make([]*model.URLStorageRecord, 0)

	q := `
		SELECT us.original_url, us.short_id, au.user_uuid, us.is_deleted 
		FROM url_storage us 
		JOIN auth_user au ON au.id = us.user_id 
		WHERE user_uuid = $1 
		AND us.is_deleted = FALSE
	`
	rows, err := s.db.QueryContext(ctx, q, userUUID)
	if err != nil {
		return nil, fmt.Errorf("query user urls from db: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var r model.URLStorageRecord
		err = rows.Scan(&r.OrigURL, &r.ShortID, &r.UserUUID, &r.IsDeleted)
		if err != nil {
			return nil, fmt.Errorf("scan user url from db: %w", err)
		}
		urls = append(urls, &r)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("get user urls from db: %w", err)
	}
	return urls, nil
}

func (s *DBURLStorage) DeleteBatch(ctx context.Context, urls model.URLDeleteBatch) error {
	if len(urls) == 0 {
		return nil
	}
	shortIDs, userUUIDs := s.segregateBatch(urls)

	q := `
    UPDATE url_storage 
    SET is_deleted = true
    WHERE short_id = ANY($1) 
    AND user_id IN (
        SELECT id FROM auth_user WHERE user_uuid = ANY($2)
    )
    AND is_deleted = false
	`

	trx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := trx.Rollback(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				s.logger.Error("failed to rollback transaction", zap.Error(err))
			}
		}
	}()

	_, err = trx.ExecContext(ctx, q, shortIDs, userUUIDs)
	if err != nil {
		return fmt.Errorf("update `is_deleted` field for urls batch: %w", err)
	}

	if cErr := trx.Commit(); cErr != nil {
		return fmt.Errorf("commiting delete batch transaction: %w", cErr)
	}
	return nil
}

func (s *DBURLStorage) segregateBatch(urls model.URLDeleteBatch) (shortIds, userUUIDs []string) {
	shortIds = make([]string, len(urls))
	userUUIDs = make([]string, len(urls))
	for i, u := range urls {
		shortIds[i] = u.ShortID
		userUUIDs[i] = u.UserUUID
	}
	return
}
