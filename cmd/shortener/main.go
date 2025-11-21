package main

import (
	"fmt"
	"log"

	"github.com/alex-storchak/shortener/internal/application"
	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("initialize config: %v", err)
	}
	zl, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("initialize logger: %v", err)
	}
	defer func() {
		//nolint:errcheck // there isn't any good strategy to log error
		_ = zl.Sync()
	}()

	if err := run(cfg, zl); err != nil {
		zl.Error("run application", zap.Error(err))
	}
}

func run(cfg *config.Config, zl *zap.Logger) error {
	var err error
	app, err := application.NewApp(cfg, zl)
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			zl.Error("close application", zap.Error(err))
		}
	}()
	if err = app.Run(); err != nil {
		return fmt.Errorf("run application: %w", err)
	}
	return nil
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	zl, err := logger.NewLogger(&cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("initialize logger with config: %v error: %w", cfg.Logger, err)
	}
	zl.Info("logger initialized")
	return zl, nil
}
