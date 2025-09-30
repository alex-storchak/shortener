package factory

import (
	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/repository"
	"go.uber.org/zap"
)

type MemoryStorageFactory struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewMemoryStorageFactory(cfg *config.Config, logger *zap.Logger) *MemoryStorageFactory {
	return &MemoryStorageFactory{
		cfg:    cfg,
		logger: logger,
	}
}

func (f *MemoryStorageFactory) MakeURLStorage() (repository.URLStorage, error) {
	storage := repository.NewMemoryURLStorage(f.logger)
	f.logger.Info("memory url storage initialized")
	return storage, nil
}

func (f *MemoryStorageFactory) MakeUserStorage() (repository.UserStorage, error) {
	storage := repository.NewMemoryUserStorage(f.logger)
	f.logger.Info("file user storage initialized")
	return storage, nil
}
