package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/alex-storchak/shortener/internal/codec"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type APIShortenBatchProcessor interface {
	Process(ctx context.Context, items []model.BatchShortenRequestItem) ([]model.BatchShortenResponseItem, error)
}

func handleAPIShortenBatch(p APIShortenBatchProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if err := validateContentType(ct, "application/json"); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		reqItems, err := codec.Decode[[]model.BatchShortenRequestItem](r)
		if err != nil {
			l.Debug("decode json request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		respItems, err := p.Process(r.Context(), reqItems)
		if errors.Is(err, service.ErrEmptyInputURL) || errors.Is(err, service.ErrEmptyInputBatch) {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if err != nil {
			l.Error("failed to shorten batch", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = codec.Encode(w, http.StatusCreated, respItems)
		if err != nil {
			l.Error("encode json response", zap.Error(err))
			return
		}
	}
}
