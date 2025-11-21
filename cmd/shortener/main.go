package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/logger"
	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/alex-storchak/shortener/internal/repository/factory"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/teris-io/shortid"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalf("failed to run application: %v", err)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	zl, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("initialize logger: %v", err)
	}
	defer func() {
		//nolint:errcheck // there isn't any good strategy to log error
		_ = zl.Sync()
	}()

	sf, err := factory.NewStorageFactory(cfg, zl)
	if err != nil {
		return fmt.Errorf("init storage factory: %w", err)
	}

	storage, err := initURLStorage(sf)
	if err != nil {
		return fmt.Errorf("init url storage: %w", err)
	}
	defer func() {
		dErr := storage.Close()
		if dErr != nil {
			zl.Error("close storage", zap.Error(dErr))
		}
	}()

	shortener, err := initShortener(storage, zl)
	if err != nil {
		return fmt.Errorf("init shortener: %w", err)
	}

	router, err := initRouter(cfg, shortener, sf, zl)
	if err != nil {
		return fmt.Errorf("init router: %w", err)
	}

	handler.Serve(ctx, cfg.Server, zl, router)
	return nil
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	zl, err := logger.New(&cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("new logger; config: %v; error: %w", cfg.Logger, err)
	}
	zl.Info("logger initialized")
	return zl, nil
}

func initURLStorage(sf factory.StorageFactory) (repository.URLStorage, error) {
	storage, err := sf.MakeURLStorage()
	if err != nil {
		return nil, fmt.Errorf("make storage: %w", err)
	}
	return storage, nil
}

func initShortIDGenerator() (*service.ShortIDGenerator, error) {
	generator, err := shortid.New(1, shortid.DefaultABC, 1)
	if err != nil {
		return nil, fmt.Errorf("instantiate shortid generator: %w", err)
	}
	return service.NewShortIDGenerator(generator), nil
}

func initShortener(s repository.URLStorage, zl *zap.Logger) (*service.Shortener, error) {
	g, err := initShortIDGenerator()
	if err != nil {
		return nil, fmt.Errorf("instantiate shortid generator: %w", err)
	}
	zl.Info("shortener initialized")
	return service.NewShortener(g, s, zl), nil
}

func initRouter(
	cfg *config.Config,
	sh service.PingableURLShortener,
	sf factory.StorageFactory,
	zl *zap.Logger,
) (http.Handler, error) {
	us, err := sf.MakeUserStorage()
	if err != nil {
		return nil, fmt.Errorf("make user storage: %w", err)
	}
	as := service.NewAuthService(zl, us, &cfg.Auth)
	um := repository.NewUserManager(zl, us)
	authMWService := service.NewAuthMiddlewareService(as, um, &cfg.Auth)

	shortenProc := service.NewMainPageService(cfg.Handler.BaseURL, sh, zl)

	shortURLProc := service.NewShortURLService(sh, zl)

	pingProc := service.NewPingService(sh, zl)

	shortenRequestDecoder := service.JSONShortenRequestDecoder{}
	apiShortenProc := service.NewShortenService(cfg.Handler.BaseURL, sh, shortenRequestDecoder, zl)

	enc := service.JSONEncoder{}

	shortenDec := service.JSONShortenBatchRequestDecoder{}
	apiShortenBatchProc := service.NewShortenBatchService(cfg.Handler.BaseURL, sh, shortenDec, zl)

	deleteDec := service.JSONDeleteBatchRequestDecoder{}
	apiUserURLsProc := service.NewAPIUserURLsService(cfg.Handler.BaseURL, sh, deleteDec, zl)

	return handler.NewRouter(
		zl,
		cfg,
		authMWService,
		shortenProc,
		shortURLProc,
		pingProc,
		apiShortenProc,
		enc,
		apiShortenBatchProc,
		apiUserURLsProc,
	), nil
}
