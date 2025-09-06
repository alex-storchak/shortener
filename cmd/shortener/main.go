package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/logger"
	"github.com/alex-storchak/shortener/internal/middleware"
	"github.com/alex-storchak/shortener/internal/repository"
	repoCfg "github.com/alex-storchak/shortener/internal/repository/config"
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
	cfg, err := initConfig()
	if err != nil {
		return err
	}

	zLogger, err := initLogger(cfg)
	if err != nil {
		return err
	}
	defer zLogger.Sync()

	shortIDGenerator, err := initShortIDGenerator()
	if err != nil {
		zLogger.Error("Failed to instantiate shortid generator", zap.Error(err), zap.String("package", "main"))
		return err
	}

	urlStorage, err := initURLStorage(cfg, zLogger)
	if err != nil {
		zLogger.Error("Failed to instantiate url storage", zap.Error(err), zap.String("package", "main"))
		return err
	}
	defer urlStorage.Close()

	shortener := service.NewShortener(shortIDGenerator, urlStorage, zLogger)

	handlers := initHandlers(cfg, zLogger, shortener)
	middlewares := initMiddlewares(zLogger)

	return handler.Serve(&cfg.Handler, handlers, middlewares)
}

func initConfig() (*config.Config, error) {
	cfg, err := config.ParseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	zLogger, err := logger.GetInstance(&cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	zLogger.Info("logger initialized")
	return zLogger, nil
}

func initShortIDGenerator() (*service.ShortIDGenerator, error) {
	generator, err := shortid.New(1, shortid.DefaultABC, 1)
	if err != nil {
		return nil, errors.New("failed to instantiate shortid generator")
	}
	return service.NewShortIDGenerator(generator), nil
}

func initURLStorage(cfg *config.Config, zLogger *zap.Logger) (*repository.FileURLStorage, error) {
	fm := repository.NewFileManager(cfg.Repository.FileStoragePath, repoCfg.DefaultFileStoragePath, zLogger)
	frp := repository.FileRecordParser{}
	fs := repository.NewFileScanner(zLogger, frp)
	um := repository.NewUUIDManager(zLogger)
	urlStorage, err := repository.NewFileURLStorage(zLogger, fm, fs, um)
	if err != nil {
		return nil, errors.New("failed to instantiate url storage")
	}
	return urlStorage, nil
}

func initHandlers(cfg *config.Config, zLogger *zap.Logger, shortener service.IShortener) *handler.Handlers {
	shortenCore := service.NewShortenCore(shortener, cfg.Handler.BaseURL, zLogger)

	mainPageService := service.NewMainPageService(shortenCore, zLogger)
	mainPageHandler := handler.NewMainPageHandler(mainPageService, zLogger)

	shortURLService := service.NewShortURLService(shortener, zLogger)
	shortURLHandler := handler.NewShortURLHandler(shortURLService, zLogger)

	jsonDecoder := service.JSONRequestDecoder{}
	apiShortenService := service.NewAPIShortenService(shortenCore, jsonDecoder, zLogger)
	jsonEncoder := service.JSONEncoder{}
	apiShortenHandler := handler.NewAPIShortenHandler(apiShortenService, jsonEncoder, zLogger)

	return &handler.Handlers{
		mainPageHandler,
		shortURLHandler,
		apiShortenHandler,
	}
}

func initMiddlewares(zLogger *zap.Logger) *handler.Middlewares {
	return &handler.Middlewares{
		middleware.RequestLogger(zLogger),
		middleware.GzipMiddleware(zLogger),
	}
}
