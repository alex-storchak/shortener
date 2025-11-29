package handler

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/middleware"
)

// NewRouter creates and configures the HTTP router with all application routes and middleware.
// It sets up the complete routing structure including authentication, logging, compression,
// and all URL shortening endpoints.
//
// Parameters:
//   - l: Structured logger for logging operations
//   - cfg: Application configuration containing server settings and auth configuration
//   - userResolver: Service for resolving and validating user authentication
//   - shortenProc: Processor for plain text URL shortening requests
//   - shortURLProc: Processor for expanding short URLs to original URLs
//   - pingProc: Processor for health check requests
//   - apiShortenProc: Processor for JSON API URL shortening requests
//   - apiShortenBatchProc: Processor for batch URL shortening operations
//   - apiUserURLsProc: Processor for user-specific URL management operations
//   - eventPublisher: Publisher for audit events tracking system actions
//
// Returns:
//   - http.Handler: Configured HTTP router with all middleware and routes
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

// Serve starts the HTTP server and manages its lifecycle.
// It begins listening to the configured address and gracefully handles server shutdown
// when the context is canceled.
//
// Parameters:
//   - ctx: Context for managing server lifecycle and graceful shutdown
//   - cfg: Server configuration containing address and shutdown timeout
//   - logger: Structured logger for logging operations
//   - router: HTTP handler with all configured routes and middleware
//
// Behavior:
//   - Starts the HTTP server in a goroutine
//   - Listens for context cancellation to initiate graceful shutdown
//   - Uses a timeout for graceful shutdown to prevent hanging connections
//   - Logs server startup and shutdown events
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
