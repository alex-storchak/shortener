package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

type ShortenProcessor interface {
	Process(ctx context.Context, body []byte) (shortURL, userUUID string, err error)
}

type AuditEventPublisher interface {
	Publish(event model.AuditEvent)
}

func handleShorten(p ShortenProcessor, l *zap.Logger, ep AuditEventPublisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		shortURL, userUUID, err := p.Process(r.Context(), body)
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
			return
		}

		ep.Publish(model.AuditEvent{
			TS:      time.Now().Unix(),
			Action:  model.AuditActionShorten,
			UserID:  userUUID,
			OrigURL: string(body),
		})
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
