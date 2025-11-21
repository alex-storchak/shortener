package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type APIShortenProcessor interface {
	Process(ctx context.Context, r io.Reader) (*model.ShortenResponse, error)
}

func handleAPIShorten(p APIShortenProcessor, enc service.Encoder, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if err := validateContentType(ct, "application/json"); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resBody, err := p.Process(r.Context(), r.Body)
		if errors.Is(err, service.ErrEmptyInputURL) {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrURLAlreadyExists) {
			if err := writeAPIShortenResponse(w, enc, http.StatusConflict, resBody); err != nil {
				l.Error("failed to write response", zap.Error(err))
			}
			return
		} else if err != nil {
			l.Error("failed to shorten", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := writeAPIShortenResponse(w, enc, http.StatusCreated, resBody); err != nil {
			l.Error("failed to write response", zap.Error(err))
		}
	}
}

func writeAPIShortenResponse(
	w http.ResponseWriter,
	enc service.Encoder,
	status int,
	body *model.ShortenResponse,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := enc.Encode(w, body); err != nil {
		return fmt.Errorf("encode response body `%v`: %w", body, err)
	}
	return nil
}
