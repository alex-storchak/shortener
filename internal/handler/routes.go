package handler

import (
	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const ShortIDParam = "id"

func addRoutes(
	mux *chi.Mux,
	l *zap.Logger,
	cfg *config.Config,
	authMWService middleware.UserResolver,
	shortenProc ShortenProcessor,
	shortURLProc ExpandProcessor,
	pingProc PingProcessor,
	apiShortenProc APIShortenProcessor,
	apiShortenBatchProc APIShortenBatchProcessor,
	apiUserURLsProc APIUserURLsProcessor,
) {
	mux.Use(middleware.NewRequestLogger(l))
	mux.Use(middleware.NewAuth(l, authMWService, cfg.Auth))
	mux.Use(middleware.NewGzip(l))

	mux.Post("/", handleShorten(shortenProc, l))
	mux.Get("/{id:[a-zA-Z0-9_-]+}", handleExpand(shortURLProc, l))
	mux.Get("/ping", handlePing(pingProc, l))

	mux.Route("/api", func(mux chi.Router) {
		mux.Route("/shorten", func(mux chi.Router) {
			mux.Post("/", handleAPIShorten(apiShortenProc, l))
			mux.Post("/batch", handleAPIShortenBatch(apiShortenBatchProc, l))
		})

		mux.Route("/user/urls", func(mux chi.Router) {
			mux.Get("/", handleGetUserURLs(apiUserURLsProc, l))
			mux.Delete("/", handleDeleteUserURLs(apiUserURLsProc, l))
		})
	})
}
