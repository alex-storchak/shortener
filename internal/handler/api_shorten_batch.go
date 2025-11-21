package handler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type APIShortenBatchProcessor interface {
	Process(ctx context.Context, r io.Reader) ([]model.BatchShortenResponseItem, error)
}

func handleAPIShortenBatch(p APIShortenBatchProcessor, enc service.Encoder, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if err := validateContentType(ct, "application/json"); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		respItems, err := p.Process(r.Context(), r.Body)
		if errors.Is(err, service.ErrEmptyInputURL) || errors.Is(err, service.ErrEmptyInputBatch) {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if err != nil {
			l.Error("failed to shorten batch", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := enc.Encode(w, respItems); err != nil {
			l.Error("failed to encode response", zap.Error(err))
			return
		}
	}
}
