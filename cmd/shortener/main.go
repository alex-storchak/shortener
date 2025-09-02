package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/logger"
	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/teris-io/shortid"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run() error {
	cfg, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// initialize logger
	zLogger, err := logger.GetInstance(&cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer zLogger.Sync()
	zLogger.Info("logger initialized")

	// initialize shortener
	generator, err := shortid.New(1, shortid.DefaultABC, 1)
	if err != nil {
		zLogger.Error("Failed to instantiate shortid generator",
			zap.Error(err),
			zap.String("package", "main"),
		)
		return errors.New("failed to instantiate shortid generator")
	}
	shortIDGenerator := service.NewShortIDGenerator(generator)
	urlStorage, err := repository.NewFileURLStorage(cfg.Repository.FileStoragePath, zLogger)
	if err != nil {
		zLogger.Error("Failed to instantiate url storage",
			zap.Error(err),
			zap.String("package", "main"),
		)
		return errors.New("failed to instantiate url storage")
	}
	defer urlStorage.Close()

	shortener := service.NewShortener(
		shortIDGenerator,
		urlStorage,
		zLogger,
	)

	return handler.Serve(&cfg.Handler, shortener, zLogger)
}
