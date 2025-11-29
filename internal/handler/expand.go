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

// ExpandProcessor defines the interface for processing URL expansion requests.
// Implementations handle the business logic of converting short URLs back to original URLs.
type ExpandProcessor interface {
	Process(ctx context.Context, shortID string) (origURL, userUUID string, err error)
}

// handleExpand creates an HTTP handler for expanding short URLs to their original URLs.
// It handles GET requests to '/{shortID}' endpoint where shortID is the URL parameter.
//
// The handler:
//   - Processes the expansion request to retrieve the original URL
//   - Returns appropriate HTTP status codes:
//   - 307 Temporary Redirect with Location header for successful expansion
//   - 404 Not Found when short ID doesn't exist
//   - 410 Gone when the URL has been deleted
//   - 500 Internal Server Error for processing failures
//   - Publishes audit events for successful URL follow actions
//
// Parameters:
//   - p: Processor implementing the URL expansion logic
//   - l: Logger for logging operations
//   - ep: Audit event publisher for recording URL follow actions
//
// Returns:
//   - HTTP handler function for the expand endpoint
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
