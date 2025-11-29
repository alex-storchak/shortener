package repository

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
)

// MemoryUserStorage provides an in-memory implementation of UserStorage.
// It stores user data in a synchronized map and is suitable for testing
// or single-instance deployments without persistence requirements.
//
// This implementation is thread-safe and uses mutex synchronization
// to handle concurrent access.
type MemoryUserStorage struct {
	logger *zap.Logger
	users  map[string]struct{}
	mu     *sync.Mutex
}

// NewMemoryUserStorage creates a new in-memory user storage instance.
//
// Parameters:
//   - logger: structured logger for logging operations
//
// Returns:
//   - *MemoryUserStorage: configured in-memory user storage
func NewMemoryUserStorage(logger *zap.Logger) *MemoryUserStorage {
	return &MemoryUserStorage{
		logger: logger,
		users:  make(map[string]struct{}, 250000),
		mu:     &sync.Mutex{},
	}
}

// Close releases resources used by the memory storage.
// For in-memory storage, this is a no-op but implements the interface.
//
// Returns:
//   - error: always returns nil
func (s *MemoryUserStorage) Close() error {
	return nil
}

// HasByUUID checks if a user with the specified UUID exists in memory storage.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used in this implementation)
//   - uuid: user UUID to check for existence
//
// Returns:
//   - bool: true if user exists in memory, false otherwise
//   - error: nil on success
func (s *MemoryUserStorage) HasByUUID(_ context.Context, uuid string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hasByUUIDUnsafe(uuid)
}

// hasByUUIDUnsafe performs the actual UUID check without locking.
// Caller must ensure proper synchronization.
func (s *MemoryUserStorage) hasByUUIDUnsafe(uuid string) (bool, error) {
	_, ok := s.users[uuid]
	return ok, nil
}

// Set stores a new user in memory storage.
// Returns an error if a user with the same UUID already exists.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used in this implementation)
//   - user: user object to store
//
// Returns:
//   - error: nil on success, or error if user already exists
func (s *MemoryUserStorage) Set(_ context.Context, user *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	has, err := s.hasByUUIDUnsafe(user.UUID)
	if err != nil {
		return fmt.Errorf("check if user exists before setting: %w", err)
	}
	if has {
		return fmt.Errorf("user with uuid %s already exists", user.UUID)
	}
	s.users[user.UUID] = struct{}{}
	return nil
}
