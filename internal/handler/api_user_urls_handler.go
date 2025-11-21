package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type UserURLsProcessor interface {
	ProcessGet(ctx context.Context) ([]model.UserURLsResponseItem, error)
	ProcessDelete(ctx context.Context, r io.Reader) error
}

func handleGetUserURLs(p UserURLsProcessor, enc service.Encoder, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respItems, err := p.ProcessGet(r.Context())
		if err != nil {
			l.Error("error getting user urls", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if len(respItems) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := enc.Encode(w, respItems); err != nil {
			l.Error("failed to encode response", zap.Error(err))
			return
		}
	}
}

func handleDeleteUserURLs(p UserURLsProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := p.ProcessDelete(r.Context(), r.Body); err != nil {
			l.Error("error deleting user urls", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}
