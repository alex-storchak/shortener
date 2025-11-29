package repository

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
)

// UserFileManager defines the interface for file management operations.
type UserFileManager interface {
	OpenForAppend(useDefault bool) (*os.File, error)
	Close() error
}

// FileUserStorage provides a file-based implementation of UserStorage.
// It persists user data to disk and restores it on initialization.
// User data is stored alongside URL records in the same file.
//
// This implementation uses file scanning to extract user UUIDs from URL records
// and maintains an in-memory index for fast lookups.
type FileUserStorage struct {
	logger   *zap.Logger
	fileMgr  UserFileManager
	fileScnr *URLFileScanner
	users    map[string]struct{}
	mu       *sync.Mutex
}

// NewFileUserStorage creates a new file-based user storage instance.
// It automatically restores existing user data from the storage file on initialization.
//
// Parameters:
//   - logger: structured logger for logging operations
//   - fm: file manager for file operations
//   - fs: file scanner for reading URL records from file
//
// Returns:
//   - *FileUserStorage: configured file-based user storage
//   - error: nil on success, or error if file restoration fails
func NewFileUserStorage(
	logger *zap.Logger,
	fm UserFileManager,
	fs *URLFileScanner,
) (*FileUserStorage, error) {
	storage := &FileUserStorage{
		logger:   logger,
		fileMgr:  fm,
		fileScnr: fs,
		mu:       &sync.Mutex{},
	}

	if err := storage.restoreFromFile(); err != nil {
		return nil, fmt.Errorf("restore storage from file: %w", err)
	}
	return storage, nil
}

// Close releases file resources used by the storage.
//
// Returns:
//   - error: nil on success, or error if file closure fails
func (s *FileUserStorage) Close() error {
	return s.fileMgr.Close()
}

// HasByUUID checks if a user with the specified UUID exists in file storage.
// Uses the in-memory index for fast lookups after initial file restoration.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used in this implementation)
//   - uuid: user UUID to check for existence
//
// Returns:
//   - bool: true if user exists, false otherwise
//   - error: nil on success
func (s *FileUserStorage) HasByUUID(_ context.Context, uuid string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hasByUUIDUnsafe(uuid)
}

// hasByUUIDUnsafe performs the actual UUID check without locking.
// Caller must ensure proper synchronization.
func (s *FileUserStorage) hasByUUIDUnsafe(uuid string) (bool, error) {
	_, ok := s.users[uuid]
	return ok, nil
}

// Set stores a new user in the file storage.
// Note: This implementation only updates the in-memory index;
// user persistence is handled through URL record storage.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used in this implementation)
//   - user: user object to store
//
// Returns:
//   - error: nil on success, or error if user already exists
func (s *FileUserStorage) Set(_ context.Context, user *model.User) error {
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

// restoreFromFile reads the storage file and rebuilds the user index
// by extracting user UUIDs from URL records.
//
// Returns:
//   - error: nil on success, or error if file operations fail
func (s *FileUserStorage) restoreFromFile() error {
	file, err := s.fileMgr.OpenForAppend(false)
	if err != nil {
		return fmt.Errorf("open requested file: %w", err)
	}

	records, err := s.fileScnr.scan(file)
	if err != nil {
		if cErr := s.fileMgr.Close(); cErr != nil {
			return fmt.Errorf("close requested file: %w", cErr)
		}
		return fmt.Errorf("scan data from requested file: %w", err)
	}

	users := make(map[string]struct{})
	for _, record := range records {
		users[record.UserUUID] = struct{}{}
	}
	s.users = users
	return nil
}
