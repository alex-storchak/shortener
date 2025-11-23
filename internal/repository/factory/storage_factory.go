package factory

import (
	"fmt"
	"strings"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/db"
	"github.com/alex-storchak/shortener/internal/file"
	"github.com/alex-storchak/shortener/internal/repository"
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
			return nil, fmt.Errorf("initialize db storage factory: %w", err)
		}
	case strings.TrimSpace(cfg.Repo.FileStoragePath) != "":
		sf, err = initFileStorageFactory(cfg, zl)
		if err != nil {
			return nil, fmt.Errorf("initialize file storage factory: %w", err)
		}
	default:
		sf = initMemoryStorageFactory(cfg, zl)
	}
	return sf, nil
}

func initDBStorageFactory(cfg *config.Config, zl *zap.Logger) (*DBStorageFactory, error) {
	d, err := db.NewDB(&cfg.DB, config.MigrationsPath, zl)
	if err != nil {
		return nil, fmt.Errorf("initialize DB: %w", err)
	}
	sf := NewDBStorageFactory(cfg, d, zl)
	zl.Info("db storage factory initialized")
	return sf, nil
}

func initFileStorageFactory(cfg *config.Config, zl *zap.Logger) (*FileStorageFactory, error) {
	fm := file.NewManager(cfg.Repo.FileStoragePath, config.DefFileStoragePath, zl)
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
