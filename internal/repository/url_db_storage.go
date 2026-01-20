package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
)

// Common database errors
var (
	// ErrDataNotFoundInDB is returned when no data is found for a query in the database.
	ErrDataNotFoundInDB = errors.New("data not found in db")
)

// DBURLStorage provides a PostgreSQL implementation of URLStorage.
// It stores URL mappings in a relational database with proper transaction support,
// concurrent access handling, and persistence between restarts.
type DBURLStorage struct {
	logger *zap.Logger
	db     *sql.DB
}

// NewDBURLStorage creates a new database URL storage instance.
//
// Parameters:
//   - logger: structured logger for logging operations
//   - db: database connection
//
// Returns:
//   - *DBURLStorage: configured database URL storage
func NewDBURLStorage(logger *zap.Logger, db *sql.DB) *DBURLStorage {
	return &DBURLStorage{
		logger: logger,
		db:     db,
	}
}

// Close closes the database connection.
//
// Returns:
//   - error: nil on success, or error if connection closure fails
func (s *DBURLStorage) Close() error {
	return s.db.Close()
}

// Ping checks if the database connection is alive and responsive.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//
// Returns:
//   - error: nil if database is accessible, or error if connection fails
func (s *DBURLStorage) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	return nil
}

// getByOriginalURL retrieves a URL record by original URL from the database.
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

// getByShortID retrieves a URL record by short ID from the database.
func (s *DBURLStorage) getByShortID(ctx context.Context, shortID string) (*model.URLStorageRecord, error) {
	q := `
		SELECT us.original_url, us.short_id, au.user_uuid, us.is_deleted 
		FROM url_storage us
		JOIN auth_user au ON au.id = us.user_id 
		WHERE us.short_id = $1
	`
	return s.getByQuery(ctx, q, shortID)
}

// getByQuery executes a database get URL query and scans the result into a URLStorageRecord.
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

// Get retrieves a URL record from the database based on search type.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - url: URL to search for
//   - searchByType: type of search (ShortURLType or OrigURLType)
//
// Returns:
//   - *model.URLStorageRecord: found record or nil if not found
//   - error: nil on success, or error if query fails or URL is deleted
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

// Set stores a single URL mapping in the database.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - origURL: original URL to be shortened
//   - shortID: generated short identifier
//   - userUUID: UUID of the user who created the mapping
//
// Returns:
//   - error: nil on success, or error if insertion fails
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

// BatchSet stores multiple URL mappings in the database within a single transaction.
// This ensures atomicity - either all records are inserted or none are.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - records: slice of URL storage records to persist
//
// Returns:
//   - error: nil on success, or error if transaction fails
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

// GetByUserUUID retrieves all non-deleted URL records for a specific user.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - userUUID: UUID of the user to retrieve URLs for
//
// Returns:
//   - []*model.URLStorageRecord: slice of URL records belonging to the user
//   - error: nil on success, or error if a query fails
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

// Count counts the amount of shortened URLs in the database.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//
// Returns:
//   - int: total amount of shortened URLs in the database
//   - error: nil on success, or database error if query fails
func (s *DBURLStorage) Count(ctx context.Context) (int, error) {
	q := "SELECT count(*) FROM url_storage"
	row := s.db.QueryRowContext(ctx, q)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("scan count urls query result row: %w", err)
	}
	return count, nil
}

// DeleteBatch marks multiple URLs as deleted in the database within a transaction.
// Only URLs belonging to the specified users are deleted.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - urls: batch of URLs to delete with user identifiers
//
// Returns:
//   - error: nil on success, or error if transaction fails
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

// segregateBatch separates URL delete batch into separate slices for short IDs and user UUIDs.
// This is used to prepare parameters for the batch delete SQL query.
func (s *DBURLStorage) segregateBatch(urls model.URLDeleteBatch) (shortIds, userUUIDs []string) {
	shortIds = make([]string, len(urls))
	userUUIDs = make([]string, len(urls))
	for i, u := range urls {
		shortIds[i] = u.ShortID
		userUUIDs[i] = u.UserUUID
	}
	return
}
