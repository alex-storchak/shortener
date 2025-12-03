package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
)

// Search type constants for URL storage operations.
const (
	ShortURLType = `shortURL`    // Search by short URL identifier
	OrigURLType  = `originalURL` // Search by original URL
)

// URLStorage defines the interface for URL data persistence operations.
// It provides methods for storing, retrieving, and managing URL mappings
// with support for different storage backends (memory, file, database).
type URLStorage interface {
	// Get retrieves a URL record based on the provided URL and search type.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//   - url: URL to search for (either short ID or original URL)
	//   - searchByType: type of search (ShortURLType or OrigURLType)
	//
	// Returns:
	//   - *model.URLStorageRecord: found record or nil if not found
	//   - error: nil on success, or storage error if operation fails
	Get(ctx context.Context, url, searchByType string) (*model.URLStorageRecord, error)

	// Set stores a new URL mapping in the storage.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//   - origURL: original URL to be shortened
	//   - shortID: generated short identifier
	//   - userUUID: UUID of the user who created the mapping
	//
	// Returns:
	//   - error: nil on success, or storage error if operation fails
	Set(ctx context.Context, origURL, shortID, userUUID string) error

	// BatchSet stores multiple URL mappings in a single operation.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//   - records: slice of URL storage records to persist
	//
	// Returns:
	//   - error: nil on success, or storage error if operation fails
	BatchSet(ctx context.Context, records []model.URLStorageRecord) error

	// Ping checks if the storage backend is accessible and responsive.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//
	// Returns:
	//   - error: nil if storage is accessible, or error if connection fails
	Ping(ctx context.Context) error

	// Close releases any resources used by the storage implementation.
	//
	// Returns:
	//   - error: nil on success, or error if cleanup fails
	Close() error

	// GetByUserUUID retrieves all URL mappings created by a specific user.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//   - userUUID: UUID of the user to retrieve URLs for
	//
	// Returns:
	//   - []*model.URLStorageRecord: slice of URL records belonging to the user
	//   - error: nil on success, or storage error if operation fails
	GetByUserUUID(ctx context.Context, userUUID string) ([]*model.URLStorageRecord, error)

	// DeleteBatch marks multiple URLs as deleted in a batch operation.
	// Only URLs belonging to the specified users can be deleted.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//   - urls: batch of URLs to delete with user identifiers
	//
	// Returns:
	//   - error: nil on success, or storage error if operation fails
	DeleteBatch(ctx context.Context, urls model.URLDeleteBatch) error
}

// DataNotFoundError represents an error when requested data is not found in storage.
// It wraps the underlying error for additional context.
type DataNotFoundError struct {
	Err error
}

// Error returns the formatted error message.
func (e *DataNotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("no data in the storage for the requested url: %v", e.Err)
	}
	return "no data in the storage for the requested url"
}

// Unwrap returns the underlying error.
func (e *DataNotFoundError) Unwrap() error {
	return e.Err
}

// NewDataNotFoundError creates a new DataNotFoundError with the specified underlying error.
func NewDataNotFoundError(err error) error {
	return &DataNotFoundError{Err: err}
}

// Common URL storage errors
var (
	// ErrDataDeleted is returned when attempting to access a URL that has been soft-deleted.
	ErrDataDeleted = errors.New("data deleted")
)
