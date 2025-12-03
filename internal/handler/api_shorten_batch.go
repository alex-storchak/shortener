package handler

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/codec"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

type APIShortenBatchProcessor interface {
	Process(ctx context.Context, items model.BatchShortenRequest) (model.BatchShortenResponse, error)
}

func handleAPIShortenBatch(p APIShortenBatchProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if err := validateContentType(ct, "application/json"); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var req model.BatchShortenRequest
		if err := codec.EasyJSONDecode(r, &req); err != nil {
			l.Debug("decode json request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		respItems, err := p.Process(r.Context(), req)
		if errors.Is(err, service.ErrEmptyInputURL) || errors.Is(err, service.ErrEmptyInputBatch) {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if err != nil {
			l.Error("failed to shorten batch", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = codec.EasyJSONEncode(w, http.StatusCreated, respItems); err != nil {
			l.Error("encode json response", zap.Error(err))
			return
		}
	}
}
