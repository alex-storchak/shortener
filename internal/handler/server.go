package handler

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func NewRouter(
	l *zap.Logger,
	cfg *config.Config,
	userResolver middleware.UserResolver,
	shortenProc ShortenProcessor,
	shortURLProc ExpandProcessor,
	pingProc PingProcessor,
	apiShortenProc APIShortenProcessor,
	apiShortenBatchProc APIShortenBatchProcessor,
	apiUserURLsProc APIUserURLsProcessor,
	eventPublisher AuditEventPublisher,
) http.Handler {
	mux := chi.NewRouter()
	addRoutes(
		mux,
		l,
		cfg,
		userResolver,
		shortenProc,
		shortURLProc,
		pingProc,
		apiShortenProc,
		apiShortenBatchProc,
		apiUserURLsProc,
		eventPublisher,
	)
	return mux
}

func Serve(
	ctx context.Context,
	cfg config.Server,
	logger *zap.Logger,
	router http.Handler,
) {
	httpServer := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}
	go func() {
		logger.Info("starting server", zap.String("server address", cfg.ServerAddr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error starting server", zap.Error(err))
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, cfg.ShutdownWaitSecsDuration)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("error shutting down http server", zap.Error(err))
		}
		logger.Info("http server closed")
	}()
	wg.Wait()
}
