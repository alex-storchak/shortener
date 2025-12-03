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

// ShortenProcessor defines the interface for processing plain text URL shortening requests.
// Implementations should handle the business logic of converting original URLs to short URLs.
type ShortenProcessor interface {
	Process(ctx context.Context, body []byte) (shortURL, userUUID string, err error)
}

// AuditEventPublisher defines the interface for publishing audit events for tracking system actions.
// Used to record events like URL shortening for auditing purposes.
type AuditEventPublisher interface {
	Publish(event model.AuditEvent)
}

// HandleShorten creates an HTTP handler for the plain text URL shortening endpoint.
// It handles POST requests with text/plain content type containing the URL to shorten.
//
// The handler:
//   - Reads the request body as the original URL
//   - Processes the shortening request
//   - Returns appropriate HTTP status codes:
//   - 400 Bad Request for empty or invalid input
//   - 409 Conflict when URL already exists (returns existing short URL)
//   - 201 Created for successful shortening
//   - 500 Internal Server Error for processing failures
//   - Publishes audit events for successful shortening operations
//
// Parameters:
//   - p: Processor implementing the shortening business logic
//   - l: Logger for logging operations
//   - ep: Audit event publisher for recording system actions
//
// Returns:
//   - HTTP handler function for the shorten endpoint
func HandleShorten(p ShortenProcessor, l *zap.Logger, ep AuditEventPublisher) http.HandlerFunc {
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

// writeResponse writes a plain text response with the specified status code and short URL.
// It sets the Content-Type header to "text/plain" and writes the short URL to the response body.
//
// Parameters:
//   - w: HTTP response writer
//   - status: HTTP status code to return
//   - shortURL: The shortened URL to send in the response body
//
// Returns: error if writing the response fails
func writeResponse(w http.ResponseWriter, status int, shortURL string) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		return fmt.Errorf("write response `%s`: %w", shortURL, err)
	}
	return nil
}
