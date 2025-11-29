package factory

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/file"
	"github.com/alex-storchak/shortener/internal/repository"
)

// FileStorageFactory implements StorageFactory for file-based storage.
// It creates storage instances that use local files as the backend with
// JSON serialization and automatic data restoration on startup.
type FileStorageFactory struct {
	fm     *file.Manager
	ufs    *repository.URLFileScanner
	logger *zap.Logger
}

// NewFileStorageFactory creates a new file storage factory instance.
//
// Parameters:
//   - fm: file manager for file operations
//   - ufs: URL file scanner for reading stored data
//   - logger: structured logger for logging operations
//
// Returns:
//   - *FileStorageFactory: configured file storage factory
func NewFileStorageFactory(
	fm *file.Manager,
	ufs *repository.URLFileScanner,
	logger *zap.Logger,
) *FileStorageFactory {
	return &FileStorageFactory{
		fm:     fm,
		ufs:    ufs,
		logger: logger,
	}
}

// MakeURLStorage creates a new file-based URL storage instance.
//
// Returns:
//   - repository.URLStorage: file-based URL storage implementation
//   - error: nil on success, or error if file restoration fails
//
// Example:
//
//	urlStorage, err := factory.MakeURLStorage()
//	if err != nil {
//	    // handle error (e.g., file permission issues)
//	}
//	defer urlStorage.Close()
func (f *FileStorageFactory) MakeURLStorage() (repository.URLStorage, error) {
	storage, err := repository.NewFileURLStorage(f.logger, f.fm, f.ufs)
	if err != nil {
		return nil, fmt.Errorf("instantiate file url storage: %w", err)
	}
	f.logger.Info("file url storage initialized")
	return storage, nil
}

// MakeUserStorage creates a new file-based user storage instance.
//
// Returns:
//   - repository.UserStorage: file-based user storage implementation
//   - error: nil on success, or error if file restoration fails
//
// Example:
//
//	userStorage, err := factory.MakeUserStorage()
//	if err != nil {
//	    // handle error (e.g., file corruption)
//	}
//	defer userStorage.Close()
func (f *FileStorageFactory) MakeUserStorage() (repository.UserStorage, error) {
	storage, err := repository.NewFileUserStorage(f.logger, f.fm, f.ufs)
	if err != nil {
		return nil, fmt.Errorf("instantiate file user storage: %w", err)
	}
	f.logger.Info("file user storage initialized")
	return storage, nil
}
