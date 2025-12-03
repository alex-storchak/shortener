package factory

import (
	"database/sql"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/repository"
	"go.uber.org/zap"
)

type DBStorageFactory struct {
	cfg    *config.Config
	db     *sql.DB
	logger *zap.Logger
}

func NewDBStorageFactory(cfg *config.Config, db *sql.DB, logger *zap.Logger) *DBStorageFactory {
	return &DBStorageFactory{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}
}

func (f *DBStorageFactory) MakeURLStorage() (repository.URLStorage, error) {
	storage := repository.NewDBURLStorage(f.logger, f.db)
	f.logger.Info("db url storage initialized")
	return storage, nil
}

func (f *DBStorageFactory) MakeUserStorage() (repository.UserStorage, error) {
	dbMgr := repository.NewUserDBManager(f.logger, f.db)
	storage := repository.NewDBUserStorage(f.logger, dbMgr)
	f.logger.Info("db user storage initialized")
	return storage, nil
}
