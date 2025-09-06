package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/handler/config"
	"github.com/alex-storchak/shortener/internal/middleware"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Serve(cfg *config.Config, shortener service.IShortener, logger *zap.Logger) error {
	h := newHandlers(shortener, cfg.BaseURL, logger)
	router := newRouter(h)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	return srv.ListenAndServe()
}

func newRouter(h *handlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(middleware.RequestLogger(h.logger))
	r.Post("/", h.MainPageHandler)
	r.Get("/{id:[a-zA-Z0-9_-]+}", h.ShortURLHandler)

	return r
}

type handlers struct {
	shortener service.IShortener
	baseURL   string
	logger    *zap.Logger
}

func newHandlers(shortener service.IShortener, baseURL string, logger *zap.Logger) *handlers {
	logger = logger.With(
		zap.String("component", "handler"),
	)
	return &handlers{
		shortener: shortener,
		baseURL:   baseURL,
		logger:    logger,
	}
}

func (h *handlers) MainPageHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		h.logger.Info("non POST request for main page")
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		h.logger.Info("request body is empty")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	shortID, err := h.shortener.Shorten(string(body))
	if err != nil {
		h.logger.Error("failed to shorten url", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	shortURL := fmt.Sprintf("%s/%s", h.baseURL, shortID)
	h.logger.Debug("shortened url", zap.String("url", shortURL))

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))
}

func (h *handlers) ShortURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		h.logger.Info("non GET request for short url")
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	shortID := chi.URLParam(req, "id")
	h.logger.Debug("short ID from request", zap.String("shortID", shortID))

	targetURL, err := h.shortener.Extract(shortID)
	if err != nil {
		if errors.Is(err, service.ErrShortenerShortIDNotFound) {
			h.logger.Info("short ID not found", zap.Error(err))
			res.WriteHeader(http.StatusNotFound)
			return
		}
		h.logger.Error("failed to extract url", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", targetURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
