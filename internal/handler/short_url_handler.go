package handler

import (
	"errors"
	"net/http"

	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ShortURLProcessor interface {
	Process(shortID string) (origURL string, err error)
}

type ShortURLHandler struct {
	srv    ShortURLProcessor
	logger *zap.Logger
}

func NewShortURLHandler(srv ShortURLProcessor, l *zap.Logger) *ShortURLHandler {
	return &ShortURLHandler{
		srv:    srv,
		logger: l,
	}
}

func (h *ShortURLHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	shortID := chi.URLParam(req, "id")

	origURL, err := h.srv.Process(shortID)
	var nfErr *repository.DataNotFoundError
	if errors.As(err, &nfErr) {
		res.WriteHeader(http.StatusNotFound)
		return
	} else if errors.Is(err, repository.ErrDataDeleted) {
		res.WriteHeader(http.StatusGone)
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
