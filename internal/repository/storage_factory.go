package repository

import (
	"fmt"
	"strings"

	"github.com/alex-storchak/shortener/internal/config"
	pkgDB "github.com/alex-storchak/shortener/internal/db"
	dbCfg "github.com/alex-storchak/shortener/internal/db/config"
	repoCfg "github.com/alex-storchak/shortener/internal/repository/config"
	"go.uber.org/zap"
)

type StorageFactory struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewStorageFactory(cfg *config.Config, zl *zap.Logger) *StorageFactory {
	return &StorageFactory{
		cfg:    cfg,
		logger: zl,
	}
}

func (f *StorageFactory) Produce() (URLStorage, error) {
	var (
		storage URLStorage
		err     error
	)
	switch {
	case strings.TrimSpace(f.cfg.DB.DSN) != "":
		storage, err = f.initDBURLStorage()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize db url storage: %w", err)
		}
	case strings.TrimSpace(f.cfg.Repository.FileStorage.Path) != "":
		storage, err = f.initFileURLStorage()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize file url storage: %w", err)
		}
	default:
		storage = f.initMemoryURLStorage()
	}
	return storage, nil
}

func (f *StorageFactory) initDBURLStorage() (URLStorage, error) {
	db, err := pkgDB.NewDB(&f.cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db: %w", err)
	}
	dbMgr, err := NewDBManager(f.logger, db, f.cfg.DB.DSN, dbCfg.MigrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db manager: %w", err)
	}
	storage := NewDBURLStorage(f.logger, dbMgr)
	f.logger.Info("db url storage initialized")
	return storage, nil
}

func (f *StorageFactory) initFileURLStorage() (URLStorage, error) {
	fm := NewFileManager(f.cfg.Repository.FileStorage.Path, repoCfg.DefaultFileStoragePath, f.logger)
	frp := FileRecordParser{}
	fs := NewFileScanner(f.logger, frp)
	um := NewUUIDManager(f.logger)
	storage, err := NewFileURLStorage(f.logger, fm, fs, um)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate file storage: %w", err)
	}
	f.logger.Info("file url storage initialized")
	return storage, nil
}

func (f *StorageFactory) initMemoryURLStorage() URLStorage {
	storage := NewMemoryURLStorage(f.logger)
	f.logger.Info("memory url storage initialized")
	return storage
}
