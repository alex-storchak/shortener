package processor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

// AuditEventPublisher defines the interface for publishing audit events for tracking system actions.
// Used to record events like URL shortening for auditing purposes.
type AuditEventPublisher interface {
	Publish(event model.AuditEvent)
}

// Shorten provides URL shortening functionality for plain text requests.
// It handles the business logic for the main '/' endpoint with text/plain content.
type Shorten struct {
	shortener service.URLShortener
	logger    *zap.Logger
	ub        ShortURLBuilder
	audit     AuditEventPublisher
}

// NewShorten creates a new Shorten processor instance.
//
// Parameters:
//   - s: URL shortener service for core shortening operations
//   - l: Structured logger for logging operations
//   - ub: URL builder for constructing complete short URLs
//   - ep: Audit event publisher for recording system actions
//
// Returns: configured Shorten processor
func NewShorten(s service.URLShortener, l *zap.Logger, ub ShortURLBuilder, ep AuditEventPublisher) *Shorten {
	return &Shorten{
		shortener: s,
		logger:    l,
		ub:        ub,
		audit:     ep,
	}
}

// Process handles the URL shortening request for plain text endpoint.
// It reads the URL from the request body and returns the shortened version.
// Also publishes audit events for successful shortening operations.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - body: request body bytes containing the original URL
//
// Returns:
//   - string: generated short URL
//   - error: nil on success, or service error if operation fails
//
// Behavior:
//   - Returns existing short URL with ErrURLAlreadyExists if URL already exists
//   - Creates new short URL for new URLs
func (s *Shorten) Process(ctx context.Context, body []byte) (string, error) {
	origURL := string(body)
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return "", fmt.Errorf("get user uuid from context: %w", err)
	}
	shortID, err := s.shortener.Shorten(ctx, userUUID, origURL)
	if errors.Is(err, service.ErrURLAlreadyExists) {
		shortURL := s.ub.Build(shortID)
		return shortURL, fmt.Errorf("tried to shorten existing url: %w", err)
	} else if err != nil {
		return "", fmt.Errorf("shorten url: %w", err)
	}

	// new url
	shortURL := s.ub.Build(shortID)

	s.audit.Publish(model.AuditEvent{
		TS:      time.Now().Unix(),
		Action:  model.AuditActionShorten,
		UserID:  userUUID,
		OrigURL: origURL,
	})

	return shortURL, nil
}
