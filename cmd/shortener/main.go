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
	urlToShortStorage := repository.NewMapURLStorage(zLogger)
	shortToURLStorage := repository.NewMapURLStorage(zLogger)
	shortener := service.NewShortener(
		shortIDGenerator,
		urlToShortStorage,
		shortToURLStorage,
		zLogger,
	)

	return handler.Serve(&cfg.Handler, shortener, zLogger)
}
