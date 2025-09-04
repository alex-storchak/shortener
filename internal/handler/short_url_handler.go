package handler

import (
	"errors"
	"net/http"

	"github.com/alex-storchak/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ShortURLHandler struct {
	shortURLSrv service.IShortURLService
	logger      *zap.Logger
}

func NewShortURLHandler(shortURLService service.IShortURLService, logger *zap.Logger) *ShortURLHandler {
	logger = logger.With(
		zap.String("component", "handler"),
		zap.String("handler", "short_url"),
	)
	return &ShortURLHandler{
		shortURLSrv: shortURLService,
		logger:      logger,
	}
}

func (h *ShortURLHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if err := validateMethod(req.Method, http.MethodGet); err != nil {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	shortID := chi.URLParam(req, "id")
	h.logger.Debug("short ID from request", zap.String("shortID", shortID))

	origURL, err := h.shortURLSrv.Expand(shortID)
	if errors.Is(err, service.ErrShortURLNotFound) {
		h.logger.Error("short url not found in storage", zap.Error(err))
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		h.logger.Error("failed to expand short url", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.Debug("original URL", zap.String("url", origURL))

	res.Header().Set("Location", origURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *ShortURLHandler) Routes() []Route {
	return []Route{
		{
			Method:  http.MethodGet,
			Pattern: "/{id:[a-zA-Z0-9_-]+}",
			Handler: h.ServeHTTP,
		},
	}
}
