package app

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/teris-io/shortid"

	"github.com/alex-storchak/shortener/internal/audit"
	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/handler/processor"
	"github.com/alex-storchak/shortener/internal/logger"
	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/alex-storchak/shortener/internal/repository/factory"
	"github.com/alex-storchak/shortener/internal/service"
)

func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	zl, err := initLogger(cfg)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	sf, err := factory.NewStorageFactory(cfg, zl)
	if err != nil {
		return fmt.Errorf("init storage factory: %w", err)
	}
	storage, err := sf.MakeURLStorage()
	if err != nil {
		return fmt.Errorf("make url storage: %w", err)
	}

	shortener, err := initShortener(storage, zl)
	if err != nil {
		return fmt.Errorf("init shortener: %w", err)
	}

	ao, err := audit.InitObservers(cfg.Audit, zl)
	if err != nil {
		return fmt.Errorf("init audit observers: %w", err)
	}
	em := audit.NewEventManager(ao, cfg.Audit, zl)

	router, err := initRouter(cfg, shortener, sf, zl, em)
	if err != nil {
		return fmt.Errorf("init router: %w", err)
	}

	handler.Serve(ctx, cfg.Server, zl, router)

	// shutdown
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.Server.ShutdownWaitSecsDuration)
	defer cancelShutdown()
	em.Close(shutdownCtx)

	if err := storage.Close(); err != nil {
		zl.Error("close storage", zap.Error(err))
	}

	//nolint:errcheck // there isn't any good strategy to log error
	_ = zl.Sync()

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

func initShortener(s repository.URLStorage, zl *zap.Logger) (*service.Shortener, error) {
	g, err := shortid.New(1, shortid.DefaultABC, 1)
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
	ep handler.AuditEventPublisher,
) (http.Handler, error) {
	us, err := sf.MakeUserStorage()
	if err != nil {
		return nil, fmt.Errorf("make user storage: %w", err)
	}
	as := service.NewAuthService(zl, us, &cfg.Auth)
	um := repository.NewUserManager(zl, us)
	ub := service.NewURLBuilder(cfg.Handler.BaseURL)
	hDeps := handler.HTTPDeps{
		Logger:              zl,
		Config:              cfg,
		UserResolver:        service.NewAuthUserResolver(as, um, &cfg.Auth),
		ShortenProc:         processor.NewShorten(sh, zl, ub),
		ExpandProc:          processor.NewExpand(sh, zl),
		PingProc:            processor.NewPing(sh, zl),
		APIShortenProc:      processor.NewAPIShorten(sh, zl, ub),
		APIShortenBatchProc: processor.NewAPIShortenBatch(sh, zl, ub),
		APIUserURLsProc:     processor.NewAPIUserURLs(sh, zl, ub),
		EventPublisher:      ep,
	}
	return handler.NewRouter(hDeps), nil
}
