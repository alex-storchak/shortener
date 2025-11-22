package handler

import (
	"context"
	"net/http"

	"github.com/alex-storchak/shortener/internal/codec"
	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type APIUserURLsProcessor interface {
	ProcessGet(ctx context.Context) ([]model.UserURLsResponseItem, error)
	ProcessDelete(ctx context.Context, shortIDs []string) error
}

func handleGetUserURLs(p APIUserURLsProcessor, l *zap.Logger) http.HandlerFunc {
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

		err = codec.Encode(w, http.StatusOK, respItems)
		if err != nil {
			l.Error("encode json response", zap.Error(err))
			return
		}
	}
}

func handleDeleteUserURLs(p APIUserURLsProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortIDs, err := codec.Decode[[]string](r)
		if err != nil {
			l.Debug("decode json request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := p.ProcessDelete(r.Context(), shortIDs); err != nil {
			l.Error("error deleting user urls", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}
