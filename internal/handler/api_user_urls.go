package handler

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/codec"
	"github.com/alex-storchak/shortener/internal/model"
)

type APIUserURLsProcessor interface {
	ProcessGet(ctx context.Context) (model.UserURLsGetResponse, error)
	ProcessDelete(ctx context.Context, shortIDs model.UserURLsDelRequest) error
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

		if err = codec.EasyJSONEncode(w, http.StatusOK, &respItems); err != nil {
			l.Error("encode json response", zap.Error(err))
			return
		}
	}
}

func handleDeleteUserURLs(p APIUserURLsProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var shortIDs model.UserURLsDelRequest
		if err := codec.EasyJSONDecode(r, &shortIDs); err != nil {
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
