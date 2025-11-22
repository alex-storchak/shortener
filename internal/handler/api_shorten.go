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

type APIShortenProcessor interface {
	Process(ctx context.Context, req model.ShortenRequest) (*model.ShortenResponse, error)
}

func handleAPIShorten(p APIShortenProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if err := validateContentType(ct, "application/json"); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		req, err := codec.Decode[model.ShortenRequest](r)
		if err != nil {
			l.Debug("decode json request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resBody, err := p.Process(r.Context(), req)
		if errors.Is(err, service.ErrEmptyInputURL) {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrURLAlreadyExists) {
			err = codec.Encode(w, http.StatusConflict, resBody)
			if err != nil {
				l.Error("conflict. encode json response", zap.Error(err))
			}
			return
		} else if err != nil {
			l.Error("failed to shorten", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = codec.Encode(w, http.StatusCreated, resBody)
		if err != nil {
			l.Error("created. encode json response", zap.Error(err))
			return
		}
	}
}
