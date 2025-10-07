package factory

import (
	"fmt"
	"strings"

	"github.com/alex-storchak/shortener/internal/config"
	pkgDB "github.com/alex-storchak/shortener/internal/db"
	dbCfg "github.com/alex-storchak/shortener/internal/db/config"
	"github.com/alex-storchak/shortener/internal/repository"
	repoCfg "github.com/alex-storchak/shortener/internal/repository/config"
	"go.uber.org/zap"
)

type StorageFactory interface {
	MakeURLStorage() (repository.URLStorage, error)
	MakeUserStorage() (repository.UserStorage, error)
}

func NewStorageFactory(cfg *config.Config, zl *zap.Logger) (StorageFactory, error) {
	var (
		sf  StorageFactory
		err error
	)
	switch {
	case strings.TrimSpace(cfg.DB.DSN) != "":
		sf, err = initDBStorageFactory(cfg, zl)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize db storage factory: %w", err)
		}
	case strings.TrimSpace(cfg.Repository.FileStorage.Path) != "":
		sf, err = initFileStorageFactory(cfg, zl)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize file storage factory: %w", err)
		}
	default:
		sf = initMemoryStorageFactory(cfg, zl)
	}
	return sf, nil
}

func initDBStorageFactory(cfg *config.Config, zl *zap.Logger) (*DBStorageFactory, error) {
	db, err := pkgDB.NewDB(&cfg.DB, dbCfg.MigrationsPath, zl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db: %w", err)
	}
	sf := NewDBStorageFactory(cfg, db, zl)
	zl.Info("db storage factory initialized")
	return sf, nil
}

func initFileStorageFactory(cfg *config.Config, zl *zap.Logger) (*FileStorageFactory, error) {
	fm := repository.NewFileManager(cfg.Repository.FileStorage.Path, repoCfg.DefaultFileStoragePath, zl)
	frp := repository.URLFileRecordParser{}
	fs := repository.NewFileScanner(zl, frp)
	sf := NewFileStorageFactory(cfg, fm, fs, zl)
	zl.Info("file storage factory initialized")
	return sf, nil
}

func initMemoryStorageFactory(cfg *config.Config, zl *zap.Logger) *MemoryStorageFactory {
	sf := NewMemoryStorageFactory(cfg, zl)
	zl.Info("memory url storage initialized")
	return sf
}
