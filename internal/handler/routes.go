package handler

import (
	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/middleware"
	"github.com/alex-storchak/shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const ShortIDParam = "id"

func addRoutes(
	mux *chi.Mux,
	logger *zap.Logger,
	cfg *config.Config,
	authMWService service.AuthMiddlewareService,
	shortenProc ShortenProcessor,
	shortURLProc ShortURLProcessor,
	pingProc PingProcessor,
	apiShortenProc APIShortenProcessor,
	enc service.Encoder,
	apiShortenBatchProc APIShortenBatchProcessor,
	apiUserURLsProc UserURLsProcessor,
) {
	mux.Use(middleware.NewRequestLogger(logger))
	mux.Use(middleware.NewAuth(logger, authMWService, cfg.Auth))
	mux.Use(middleware.NewGzip(logger))

	mux.Post("/", handleShorten(shortenProc, logger))
	mux.Get("/{id:[a-zA-Z0-9_-]+}", handleExpand(shortURLProc, logger))
	mux.Get("/ping", handlePing(pingProc, logger))

	mux.Route("/api", func(mux chi.Router) {
		mux.Route("/shorten", func(mux chi.Router) {
			mux.Post("/", handleAPIShorten(apiShortenProc, enc, logger))
			mux.Post("/batch", handleAPIShortenBatch(apiShortenBatchProc, enc, logger))
		})

		mux.Route("/user/urls", func(mux chi.Router) {
			mux.Get("/", handleGetUserURLs(apiUserURLsProc, enc, logger))
			mux.Delete("/", handleDeleteUserURLs(apiUserURLsProc, logger))
		})
	})
}
