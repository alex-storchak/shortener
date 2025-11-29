package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/repository"
)

type ExpandProcessor interface {
	Process(ctx context.Context, shortID string) (origURL, userUUID string, err error)
}

func handleExpand(p ExpandProcessor, l *zap.Logger, ep AuditEventPublisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortID := chi.URLParam(r, ShortIDParam)

		origURL, userUUID, err := p.Process(r.Context(), shortID)
		var nfErr *repository.DataNotFoundError
		if errors.As(err, &nfErr) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if errors.Is(err, repository.ErrDataDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		} else if err != nil {
			l.Error("failed to expand short url", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", origURL)
		w.WriteHeader(http.StatusTemporaryRedirect)

		ep.Publish(model.AuditEvent{
			TS:      time.Now().Unix(),
			Action:  model.AuditActionFollow,
			UserID:  userUUID,
			OrigURL: origURL,
		})
	}
}
