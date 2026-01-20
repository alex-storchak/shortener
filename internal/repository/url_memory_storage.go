package repository

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
)

// MemoryURLStorage provides an in-memory implementation of URLStorage.
// It stores URL mappings in a synchronized slice and is suitable for testing
// or single-instance deployments without persistence requirements.
//
// This implementation is thread-safe and uses mutex synchronization
// to handle concurrent access.
//
// It does not persist data between restarts.
type MemoryURLStorage struct {
	logger  *zap.Logger
	records []model.URLStorageRecord
	mu      *sync.Mutex
}

// NewMemoryURLStorage creates a new in-memory URL storage instance.
//
// Parameters:
//   - logger: structured logger for logging operations
//
// Returns:
//   - *MemoryURLStorage: configured in-memory URL storage
func NewMemoryURLStorage(logger *zap.Logger) *MemoryURLStorage {
	return &MemoryURLStorage{
		logger:  logger,
		records: make([]model.URLStorageRecord, 0, 250000),
		mu:      &sync.Mutex{},
	}
}

// Close releases resources used by the memory storage.
// For in-memory storage, this is a no-op but implements the interface.
//
// Returns:
//   - error: always returns nil
func (s *MemoryURLStorage) Close() error {
	return nil
}

// Ping always returns nil for in-memory storage as it's always available.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//
// Returns:
//   - error: always returns nil
func (s *MemoryURLStorage) Ping(_ context.Context) error {
	return nil
}

// Get retrieves a URL record from memory storage based on search type.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//   - url: URL to search for
//   - searchByType: type of search (ShortURLType or OrigURLType)
//
// Returns:
//   - *model.URLStorageRecord: found record or nil if not found
//   - error: nil on success, or ErrDataDeleted if URL is deleted
func (s *MemoryURLStorage) Get(_ context.Context, url, searchByType string) (*model.URLStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch searchByType {
	case OrigURLType:
		for _, r := range s.records {
			if r.OrigURL == url && !r.IsDeleted {
				return &r, nil
			}
		}
	case ShortURLType:
		for _, r := range s.records {
			if r.ShortID == url {
				if r.IsDeleted {
					return nil, ErrDataDeleted
				}
				return &r, nil
			}
		}
	}
	return nil, NewDataNotFoundError(nil)
}

// Set stores a single URL mapping in memory storage.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//   - origURL: original URL to be shortened
//   - shortID: generated short identifier
//   - userUUID: UUID of the user who created the mapping
//
// Returns:
//   - error: nil on success
func (s *MemoryURLStorage) Set(_ context.Context, origURL, shortID, userUUID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, model.URLStorageRecord{
		OrigURL:  origURL,
		ShortID:  shortID,
		UserUUID: userUUID,
	})
	return nil
}

// BatchSet stores multiple URL mappings in memory storage in a single operation.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//   - records: slice of URL storage records to persist
//
// Returns:
//   - error: nil on success
func (s *MemoryURLStorage) BatchSet(_ context.Context, records []model.URLStorageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, records...)
	return nil
}

// GetByUserUUID retrieves all non-deleted URL mappings for a specific user.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//   - userUUID: UUID of the user to retrieve URLs for
//
// Returns:
//   - []*model.URLStorageRecord: slice of URL records belonging to the user
//   - error: nil on success
func (s *MemoryURLStorage) GetByUserUUID(_ context.Context, userUUID string) ([]*model.URLStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := make([]*model.URLStorageRecord, 0, 50)
	for i := range s.records {
		r := &s.records[i]
		if r.UserUUID == userUUID && !r.IsDeleted {
			records = append(records, r)
		}
	}
	return records, nil
}

// Count counts the amount of shortened URLs in memory storage.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used in this implementation)
//
// Returns:
//   - int: total amount of shortened URLs in the database
//   - error: nil on success, or database error if query fails
func (s *MemoryURLStorage) Count(_ context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.records), nil
}

// DeleteBatch marks multiple URLs as deleted in memory storage.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//   - urls: batch of URLs to delete with user identifiers
//
// Returns:
//   - error: nil on success
func (s *MemoryURLStorage) DeleteBatch(_ context.Context, urls model.URLDeleteBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ProcessMemDeleteBatch(s.records, urls)
	return nil
}

// ProcessMemDeleteBatch processes URL deletion in memory by marking records as deleted.
// This function is used by both MemoryURLStorage and FileURLStorage implementations.
//
// Parameters:
//   - records: slice of URL storage records to process
//   - urls: batch of URLs to mark as deleted
func ProcessMemDeleteBatch(records []model.URLStorageRecord, urls model.URLDeleteBatch) {
	if len(urls) == 0 {
		return
	}

	deleteMap := make(map[string]map[string]bool)
	for _, u := range urls {
		if deleteMap[u.ShortID] == nil {
			deleteMap[u.ShortID] = make(map[string]bool)
		}
		deleteMap[u.ShortID][u.UserUUID] = true
	}

	for i := range records {
		r := &records[i]
		if userMap, exists := deleteMap[r.ShortID]; exists {
			if userMap[r.UserUUID] {
				r.IsDeleted = true
			}
		}
	}
}
