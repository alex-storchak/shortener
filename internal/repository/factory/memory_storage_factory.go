package factory

import (
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/repository"
)

// MemoryStorageFactory implements StorageFactory for in-memory storage.
// It creates storage instances that use Go data structures as the backend
// with no persistence between application restarts.
type MemoryStorageFactory struct {
	logger *zap.Logger
}

// NewMemoryStorageFactory creates a new memory storage factory instance.
//
// Parameters:
//   - logger: structured logger for logging operations
//
// Returns:
//   - *MemoryStorageFactory: configured memory storage factory
func NewMemoryStorageFactory(logger *zap.Logger) *MemoryStorageFactory {
	return &MemoryStorageFactory{
		logger: logger,
	}
}

// MakeURLStorage creates a new memory-based URL storage instance.
//
// Returns:
//   - repository.URLStorage: memory URL storage implementation
//   - error: always returns nil for memory storage
func (f *MemoryStorageFactory) MakeURLStorage() (repository.URLStorage, error) {
	storage := repository.NewMemoryURLStorage(f.logger)
	f.logger.Info("memory url storage initialized")
	return storage, nil
}

// MakeUserStorage creates a new memory-based user storage instance.
//
// Returns:
//   - repository.UserStorage: memory user storage implementation
//   - error: always returns nil for memory storage
func (f *MemoryStorageFactory) MakeUserStorage() (repository.UserStorage, error) {
	storage := repository.NewMemoryUserStorage(f.logger)
	f.logger.Info("file user storage initialized")
	return storage, nil
}
