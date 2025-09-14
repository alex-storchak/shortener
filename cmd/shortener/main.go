package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/alex-storchak/shortener/internal/config"
	pkgDB "github.com/alex-storchak/shortener/internal/db"
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

	db, err := initDB(cfg)
	if err != nil {
		zLogger.Error("Failed to instantiate database", zap.Error(err), zap.String("package", "main"))
		return err
	}
	defer db.Close()

	urlStorage, closer, err := initURLStorage(cfg, zLogger, db)
	if err != nil {
		zLogger.Error("Failed to instantiate url storage", zap.Error(err), zap.String("package", "main"))
		return err
	}
	if closer != nil {
		defer closer.Close()
	}

	shortener := service.NewShortener(shortIDGenerator, urlStorage, zLogger)

	handlers := initHandlers(cfg, zLogger, shortener, db)
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

func initDB(cfg *config.Config) (*sql.DB, error) {
	db, err := pkgDB.GetInstance(&cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db: %w", err)
	}
	return db, nil
}

func initShortIDGenerator() (*service.ShortIDGenerator, error) {
	generator, err := shortid.New(1, shortid.DefaultABC, 1)
	if err != nil {
		return nil, errors.New("failed to instantiate shortid generator")
	}
	return service.NewShortIDGenerator(generator), nil
}

func initURLStorage(cfg *config.Config, zLogger *zap.Logger, db *sql.DB) (repository.URLStorage, io.Closer, error) {
	if strings.TrimSpace(cfg.DB.DSN) != "" && db != nil {
		dbMgr := repository.NewDBManager(zLogger, db)
		storage := repository.NewDBURLStorage(zLogger, dbMgr)
		zLogger.Info("db url storage initialized")
		return storage, io.NopCloser(nil), nil
	}

	if strings.TrimSpace(cfg.Repository.FileStorage.Path) != "" {
		fm := repository.NewFileManager(cfg.Repository.FileStorage.Path, repoCfg.DefaultFileStoragePath, zLogger)
		frp := repository.FileRecordParser{}
		fs := repository.NewFileScanner(zLogger, frp)
		um := repository.NewUUIDManager(zLogger)
		fileStorage, err := repository.NewFileURLStorage(zLogger, fm, fs, um)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to instantiate file storage: %w", err)
		}
		zLogger.Info("file url storage initialized")
		return fileStorage, fileStorage, nil
	}

	memStorage := repository.NewMemoryURLStorage(zLogger)
	zLogger.Info("memory url storage initialized")
	return memStorage, io.NopCloser(nil), nil
}

func initHandlers(
	cfg *config.Config,
	zLogger *zap.Logger,
	shortener service.IShortener,
	db *sql.DB,
) *handler.Handlers {
	shortenCore := service.NewShortenCore(shortener, cfg.Handler.BaseURL, zLogger)

	mainPageService := service.NewMainPageService(shortenCore, zLogger)
	mainPageHandler := handler.NewMainPageHandler(mainPageService, zLogger)

	shortURLService := service.NewShortURLService(shortener, zLogger)
	shortURLHandler := handler.NewShortURLHandler(shortURLService, zLogger)

	jsonDecoder := service.JSONRequestDecoder{}
	apiShortenService := service.NewAPIShortenService(shortenCore, jsonDecoder, zLogger)
	jsonEncoder := service.JSONEncoder{}
	apiShortenHandler := handler.NewAPIShortenHandler(apiShortenService, jsonEncoder, zLogger)

	pingDBService := service.NewPingDBService(db, zLogger)
	pingDBHandler := handler.NewPingDBHandler(pingDBService, zLogger)

	return &handler.Handlers{
		mainPageHandler,
		shortURLHandler,
		apiShortenHandler,
		pingDBHandler,
	}
}

func initMiddlewares(zLogger *zap.Logger) *handler.Middlewares {
	return &handler.Middlewares{
		middleware.RequestLogger(zLogger),
		middleware.GzipMiddleware(zLogger),
	}
}
