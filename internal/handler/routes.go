package handler

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/middleware"
)

// ShortIDParam defines the URL parameter name for short URL identifiers.
// Used in routes like '/{id}' where 'id' is the short URL identifier.
const ShortIDParam = "id"

// addRoutes configures all HTTP routes and middleware for the application.
// It sets up the complete routing hierarchy including:
// - Global middleware (logging, compression, recovery)
// - Debug endpoints for profiling
// - Application routes with authentication
// - API endpoints for URL operations
//
// Parameters:
//   - mux: Chi router instance to configure
//   - l: Structured logger for logging operations
//   - cfg: Application configuration
//   - userResolver: Service for user authentication resolution
//   - shorten: Processor for plain text URL shortening
//   - expand: Processor for URL expansion
//   - ping: Processor for health checks
//   - apiShorten: Processor for JSON API shortening
//   - apiShortenBatch: Processor for batch URL shortening
//   - apiUserURLs: Processor for user URL management
//   - eventPublisher: Publisher for audit events
func addRoutes(
	mux *chi.Mux,
	l *zap.Logger,
	cfg *config.Config,
	userResolver middleware.UserResolver,
	shorten ShortenProcessor,
	expand ExpandProcessor,
	ping PingProcessor,
	apiShorten APIShortenProcessor,
	apiShortenBatch APIShortenBatchProcessor,
	apiUserURLs APIUserURLsProcessor,
	eventPublisher AuditEventPublisher,
) {
	mux.Use(middleware.NewRequestLogger(l))
	mux.Use(middleware.NewGzip(l))
	mux.Use(chimw.Recoverer)

	// pprof endpoints
	mux.Route("/debug", func(mux chi.Router) {
		mux.Mount("/", chimw.Profiler())
	})

	// app endpoints
	mux.Route("/", func(mux chi.Router) {
		mux.Use(middleware.NewAuth(l, userResolver, cfg.Auth))

		mux.Post("/", handleShorten(shorten, l, eventPublisher))
		mux.Get("/{id:[a-zA-Z0-9_-]+}", handleExpand(expand, l, eventPublisher))
		mux.Get("/ping", handlePing(ping, l))

		mux.Route("/api", func(mux chi.Router) {
			mux.Route("/shorten", func(mux chi.Router) {
				mux.Post("/", handleAPIShorten(apiShorten, l, eventPublisher))
				mux.Post("/batch", handleAPIShortenBatch(apiShortenBatch, l))
			})

			mux.Route("/user/urls", func(mux chi.Router) {
				mux.Get("/", handleGetUserURLs(apiUserURLs, l))
				mux.Delete("/", handleDeleteUserURLs(apiUserURLs, l))
			})
		})
	})
}
