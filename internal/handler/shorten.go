package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type ShortenProcessor interface {
	Process(ctx context.Context, body []byte) (shortURL string, err error)
}

func handleShorten(p ShortenProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		shortURL, err := p.Process(r.Context(), body)
		if errors.Is(err, service.ErrEmptyInputURL) {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrURLAlreadyExists) {
			if err := writeResponse(w, http.StatusConflict, shortURL); err != nil {
				l.Error("failed to write response (status conflict) for main page request", zap.Error(err))
			}
			return
		} else if err != nil {
			l.Error("failed to process main page request", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := writeResponse(w, http.StatusCreated, shortURL); err != nil {
			l.Error("failed to write response (status created) for main page request", zap.Error(err))
		}
	}
}

func writeResponse(w http.ResponseWriter, status int, shortURL string) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		return fmt.Errorf("write response `%s`: %w", shortURL, err)
	}
	return nil
}
