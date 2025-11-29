package factory

import (
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/db"
	"github.com/alex-storchak/shortener/internal/file"
	"github.com/alex-storchak/shortener/internal/repository"
)

// StorageFactory defines the interface for creating storage instances.
// It provides methods for creating both URL and user storage implementations
// with consistent configuration and initialization.
type StorageFactory interface {
	// MakeURLStorage creates and initializes a URL storage instance.
	//
	// Returns:
	//   - repository.URLStorage: configured URL storage implementation
	//   - error: nil on success, or error if initialization fails
	MakeURLStorage() (repository.URLStorage, error)

	// MakeUserStorage creates and initializes a user storage instance.
	//
	// Returns:
	//   - repository.UserStorage: configured user storage implementation
	//   - error: nil on success, or error if initialization fails
	MakeUserStorage() (repository.UserStorage, error)
}

// NewStorageFactory creates the appropriate storage factory based on configuration.
// The factory selection follows a priority order:
//   - Database storage if DSN is configured
//   - File storage if file storage path is configured
//   - Memory storage as fallback
//
// Parameters:
//   - cfg: application configuration containing storage settings
//   - zl: structured logger for logging operations
//
// Returns:
//   - StorageFactory: appropriate storage factory instance
//   - error: nil on success, or error if factory initialization fails
//
// Example:
//
//	factory, err := NewStorageFactory(cfg, logger)
//	if err != nil {
//	    // handle error
//	}
//	urlStorage, err := factory.MakeURLStorage()
//	userStorage, err := factory.MakeUserStorage()
func NewStorageFactory(cfg *config.Config, zl *zap.Logger) (StorageFactory, error) {
	var (
		sf  StorageFactory
		err error
	)
	switch {
	case strings.TrimSpace(cfg.DB.DSN) != "":
		sf, err = initDBStorageFactory(cfg, zl)
		if err != nil {
			return nil, fmt.Errorf("initialize db storage factory: %w", err)
		}
	case strings.TrimSpace(cfg.Repo.FileStoragePath) != "":
		sf, err = initFileStorageFactory(cfg, zl)
		if err != nil {
			return nil, fmt.Errorf("initialize file storage factory: %w", err)
		}
	default:
		sf = initMemoryStorageFactory(zl)
	}
	return sf, nil
}

// initDBStorageFactory initializes a database storage factory with database connection.
func initDBStorageFactory(cfg *config.Config, zl *zap.Logger) (*DBStorageFactory, error) {
	d, err := db.NewDB(&cfg.DB, config.MigrationsPath, zl)
	if err != nil {
		return nil, fmt.Errorf("initialize DB: %w", err)
	}
	sf := NewDBStorageFactory(d, zl)
	zl.Info("db storage factory initialized")
	return sf, nil
}

// initFileStorageFactory initializes a file storage factory with file manager and scanner.
func initFileStorageFactory(cfg *config.Config, zl *zap.Logger) (*FileStorageFactory, error) {
	fm := file.NewManager(cfg.Repo.FileStoragePath, config.DefFileStoragePath, zl)
	frp := repository.URLFileRecordParser{}
	fs := repository.NewFileScanner(zl, frp)
	sf := NewFileStorageFactory(fm, fs, zl)
	zl.Info("file storage factory initialized")
	return sf, nil
}

// initMemoryStorageFactory initializes a memory storage factory without external dependencies.
func initMemoryStorageFactory(zl *zap.Logger) *MemoryStorageFactory {
	sf := NewMemoryStorageFactory(zl)
	zl.Info("memory url storage initialized")
	return sf
}
