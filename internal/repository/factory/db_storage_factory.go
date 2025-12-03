package factory

import (
	"database/sql"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/repository"
)

// DBStorageFactory implements StorageFactory for database-based storage.
// It creates storage instances that use PostgreSQL as the backend with
// connection pooling and transaction support.
type DBStorageFactory struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewDBStorageFactory creates a new database storage factory instance.
//
// Parameters:
//   - db: established database connection pool
//   - logger: structured logger for logging operations
//
// Returns:
//   - *DBStorageFactory: configured database storage factory
func NewDBStorageFactory(db *sql.DB, logger *zap.Logger) *DBStorageFactory {
	return &DBStorageFactory{
		db:     db,
		logger: logger,
	}
}

// MakeURLStorage creates a new database-based URL storage instance.
//
// Returns:
//   - repository.URLStorage: database URL storage implementation
//   - error: always returns nil for database storage
func (f *DBStorageFactory) MakeURLStorage() (repository.URLStorage, error) {
	storage := repository.NewDBURLStorage(f.logger, f.db)
	f.logger.Info("db url storage initialized")
	return storage, nil
}

// MakeUserStorage creates a new database-based user storage instance.
//
// Returns:
//   - repository.UserStorage: database user storage implementation
//   - error: always returns nil for database storage
func (f *DBStorageFactory) MakeUserStorage() (repository.UserStorage, error) {
	storage := repository.NewDBUserStorage(f.logger, f.db)
	f.logger.Info("db user storage initialized")
	return storage, nil
}
