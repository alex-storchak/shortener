package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alex-storchak/shortener/internal/application"
	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("failed to initialize config: %v", err)
	}
	zl, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		if sErr := zl.Sync(); sErr != nil {
			fmt.Fprintf(os.Stderr, "logger sync error: %v\n", sErr)
		}
	}()

	if err := run(cfg, zl); err != nil {
		zl.Error("failed to run application", zap.Error(err))
	}
}

func run(cfg *config.Config, zl *zap.Logger) error {
	var err error
	app, err := application.NewApp(cfg, zl)
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			zl.Error("failed to close application", zap.Error(err))
		}
	}()
	if err = app.Run(); err != nil {
		return fmt.Errorf("application runtime error: %w", err)
	}
	return nil
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	zl, err := logger.NewLogger(&cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger with config: %v error: %w", cfg.Logger, err)
	}
	zl.Info("logger initialized")
	return zl, nil
}
