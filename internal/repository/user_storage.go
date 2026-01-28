package repository

import (
	"context"

	"github.com/alex-storchak/shortener/internal/model"
)

// UserStorage defines the interface for user data persistence operations.
// It provides methods for checking user existence and storing user data.
// Implementations can use different storage backends (memory, file, database).
type UserStorage interface {
	// Count counts a total amount of users in storage.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//
	// Returns:
	//   - int: total amount of users in storage
	//   - error: nil on success, or storage error if operation fails
	Count(ctx context.Context) (int, error)

	// HasByUUID checks if a user with the specified UUID exists in storage.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//   - uuid: user UUID to check for existence
	//
	// Returns:
	//   - bool: true if user exists, false otherwise
	//   - error: nil on success, or storage error if operation fails
	HasByUUID(ctx context.Context, uuid string) (bool, error)

	// Set stores a new user in the storage.
	// Implementations should ensure user UUID uniqueness.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//   - user: user object to store
	//
	// Returns:
	//   - error: nil on success, or error if user already exists or storage operation fails
	Set(ctx context.Context, user *model.User) error

	// Close releases any resources used by the storage implementation.
	// This should be called when the storage is no longer needed.
	//
	// Returns:
	//   - error: nil on success, or error if cleanup fails
	Close() error
}
