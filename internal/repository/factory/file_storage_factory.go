package factory

import (
	"fmt"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/repository"
	"go.uber.org/zap"
)

type FileStorageFactory struct {
	cfg    *config.Config
	fm     *repository.FileManager
	ufs    *repository.URLFileScanner
	logger *zap.Logger
}

func NewFileStorageFactory(
	cfg *config.Config,
	fm *repository.FileManager,
	ufs *repository.URLFileScanner,
	logger *zap.Logger,
) *FileStorageFactory {
	return &FileStorageFactory{
		cfg:    cfg,
		fm:     fm,
		ufs:    ufs,
		logger: logger,
	}
}

func (f *FileStorageFactory) MakeURLStorage() (repository.URLStorage, error) {
	storage, err := repository.NewFileURLStorage(f.logger, f.fm, f.ufs)
	if err != nil {
		return nil, fmt.Errorf("instantiate file url storage: %w", err)
	}
	f.logger.Info("file url storage initialized")
	return storage, nil
}

func (f *FileStorageFactory) MakeUserStorage() (repository.UserStorage, error) {
	storage, err := repository.NewFileUserStorage(f.logger, f.fm, f.ufs)
	if err != nil {
		return nil, fmt.Errorf("instantiate file user storage: %w", err)
	}
	f.logger.Info("file user storage initialized")
	return storage, nil
}
