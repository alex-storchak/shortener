package handler

import (
	"errors"
	"net/http"

	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ShortURLHandler struct {
	shortURLSrv service.IShortURLService
	logger      *zap.Logger
}

func NewShortURLHandler(shortURLService service.IShortURLService, logger *zap.Logger) *ShortURLHandler {
	return &ShortURLHandler{
		shortURLSrv: shortURLService,
		logger:      logger,
	}
}

func (h *ShortURLHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	shortID := chi.URLParam(req, "id")

	origURL, err := h.shortURLSrv.Expand(shortID)
	var nfErr *repository.DataNotFoundError
	if errors.As(err, &nfErr) {
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		h.logger.Error("failed to expand short url", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

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
