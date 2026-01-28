package handler

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

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
//   - h: HTTPDeps containing all dependencies required for HTTP server initialization
func addRoutes(
	mux *chi.Mux,
	h HTTPDeps,
) {
	mux.Use(middleware.NewRequestLogger(h.Logger))
	mux.Use(middleware.NewGzip(h.Logger))
	mux.Use(chimw.Recoverer)

	// pprof endpoints
	mux.Route("/debug", func(mux chi.Router) {
		mux.Mount("/", chimw.Profiler())
	})

	// app endpoints
	mux.Route("/", func(mux chi.Router) {
		mux.Use(middleware.NewAuth(h.Logger, h.UserResolver, h.Config.Auth))

		mux.Post("/", HandleShorten(h.ShortenProc, h.Logger, h.EventPublisher))
		mux.Get("/{id:[a-zA-Z0-9_-]+}", HandleExpand(h.ExpandProc, h.Logger, h.EventPublisher))
		mux.Get("/ping", HandlePing(h.PingProc, h.Logger))

		mux.Route("/api", func(mux chi.Router) {
			mux.Route("/shorten", func(mux chi.Router) {
				mux.Post("/", HandleAPIShorten(h.APIShortenProc, h.Logger, h.EventPublisher))
				mux.Post("/batch", HandleAPIShortenBatch(h.APIShortenBatchProc, h.Logger))
			})

			mux.Route("/user/urls", func(mux chi.Router) {
				mux.Get("/", HandleGetUserURLs(h.APIUserURLsProc, h.Logger))
				mux.Delete("/", HandleDeleteUserURLs(h.APIUserURLsProc, h.Logger))
			})

			mux.Route("/internal", func(mux chi.Router) {
				mux.Use(middleware.NewTrustedSubnet(h.Logger, h.Config.Server.TrustedSubnet))

				mux.Get("/stats", HandleStats(h.APIInternalProc, h.Logger))
			})
		})
	})
}
