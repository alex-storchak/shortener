package app

import (
	"context"
	"fmt"

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

func Run(
	ctx context.Context,
	args []string,
	lookupEnv func(string) (string, bool),
) error {
	cfg, err := config.Load(args[1:], lookupEnv)
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

	deps, err := initServerDeps(cfg, shortener, sf, zl, em)
	if err != nil {
		return fmt.Errorf("init server dependencies: %w", err)
	}
	router := handler.NewRouter(deps)

	httpServer, err := handler.Serve(cfg.Server, zl, router)
	if err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	grpcServer, err := handler.ServeGRPC(deps)
	if err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}

	<-ctx.Done()

	// shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownWaitSecsDuration)
	defer cancel()

	// shutdown grpc server
	grpcStopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopped)
	}()

	select {
	case <-grpcStopped:
		zl.Info("grpc server stopped gracefully")
	case <-shutdownCtx.Done():
		zl.Warn("grpc shutdown forced by timeout")
		grpcServer.Stop()
	}

	// shutdown http server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zl.Error("http server shutdown error", zap.Error(err))
	}
	zl.Info("http server closed")

	em.Close(shutdownCtx)

	if err := storage.Close(); err != nil {
		zl.Error("failed to close storage", zap.Error(err))
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

func initServerDeps(
	cfg *config.Config,
	sh service.PingableURLShortener,
	sf factory.StorageFactory,
	zl *zap.Logger,
	ep processor.AuditEventPublisher,
) (*handler.ServerDeps, error) {
	us, err := sf.MakeUserStorage()
	if err != nil {
		return nil, fmt.Errorf("make user storage: %w", err)
	}
	as := service.NewAuthService(zl, us, &cfg.Auth)
	um := repository.NewUserManager(zl, us)
	ub := service.NewURLBuilder(cfg.Handler.BaseURL)
	hDeps := handler.ServerDeps{
		Logger:              zl,
		Config:              cfg,
		HTTPUserResolver:    service.NewAuthUserResolver(as, um, &cfg.Auth),
		GRPCUserResolver:    service.NewAuthUserResolver(as, um, &cfg.Auth),
		ShortenProc:         processor.NewShorten(sh, zl, ub, ep),
		ExpandProc:          processor.NewExpand(sh, zl, ep),
		PingProc:            processor.NewPing(sh, zl),
		APIShortenProc:      processor.NewAPIShorten(sh, zl, ub, ep),
		APIShortenBatchProc: processor.NewAPIShortenBatch(sh, zl, ub),
		APIUserURLsProc:     processor.NewAPIUserURLs(sh, zl, ub),
		APIInternalProc:     processor.NewAPIInternal(us, sh),
	}
	return &hDeps, nil
}
