package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
)

// DBUserStorage provides a database implementation of UserStorage using PostgreSQL.
// It stores user data in a relational database with proper transaction support
// and concurrent access handling.
type DBUserStorage struct {
	logger *zap.Logger
	db     *sql.DB
}

// NewDBUserStorage creates a new database user storage instance.
//
// Parameters:
//   - logger: structured logger for logging operations
//   - db: database connection
//
// Returns:
//   - *DBUserStorage: configured database user storage
func NewDBUserStorage(logger *zap.Logger, db *sql.DB) *DBUserStorage {
	return &DBUserStorage{
		logger: logger,
		db:     db,
	}
}

// Close closes the database connection.
//
// Returns:
//   - error: nil on success, or error if connection closure fails
func (s *DBUserStorage) Close() error {
	return s.db.Close()
}

// HasByUUID checks if a user with the specified UUID exists in the database.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - uuid: user UUID to check for existence
//
// Returns:
//   - bool: true if user exists in database, false otherwise
//   - error: nil on success, or database error if query fails
func (s *DBUserStorage) HasByUUID(ctx context.Context, uuid string) (bool, error) {
	q := "SELECT id FROM auth_user WHERE user_uuid = $1"
	row := s.db.QueryRowContext(ctx, q, uuid)

	var id int32
	err := row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("scan check user exists query result row: %w", err)
	}
	return true, nil
}

// Set stores a new user in the database.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - user: user object to store
//
// Returns:
//   - error: nil on success, or error if the user already exists or insertion fails
func (s *DBUserStorage) Set(ctx context.Context, user *model.User) error {
	q := "INSERT INTO auth_user (user_uuid) VALUES ($1)"
	_, err := s.db.ExecContext(ctx, q, user.UUID)
	if err != nil {
		return fmt.Errorf("persist user (%s) to db: %w", user.UUID, err)
	}
	return nil
}
