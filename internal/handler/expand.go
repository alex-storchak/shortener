package handler

import (
	"errors"
	"net/http"

	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ExpandProcessor interface {
	Process(shortID string) (origURL string, err error)
}

func handleExpand(p ExpandProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortID := chi.URLParam(r, ShortIDParam)

		origURL, err := p.Process(shortID)
		var nfErr *repository.DataNotFoundError
		if errors.As(err, &nfErr) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if errors.Is(err, repository.ErrDataDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		} else if err != nil {
			l.Error("failed to expand short url", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", origURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
