package processor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/service"
)

// Expand provides URL expansion functionality for retrieving original URLs from short identifiers.
// It handles the business logic for the '/{shortID}' endpoint.
type Expand struct {
	shortener service.URLShortener
	logger    *zap.Logger
}

// NewExpand creates a new Expand processor instance.
//
// Parameters:
//   - shortener: URL shortener service for URL extraction
//   - logger: Structured logger for logging operations
//
// Returns: configured Expand processor
func NewExpand(shortener service.URLShortener, logger *zap.Logger) *Expand {
	return &Expand{
		shortener: shortener,
		logger:    logger,
	}
}

// Process handles the URL expansion request to retrieve original URL from short ID.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - shortID: short identifier to expand
//
// Returns:
//   - string: original URL associated with the short ID
//   - string: user UUID for audit purposes (empty if not authenticated)
//   - error: nil on success, or storage error if URL not found or deleted
func (s *Expand) Process(ctx context.Context, shortID string) (string, string, error) {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		s.logger.Debug("failed to get user uuid from context", zap.Error(err))
		userUUID = ""
	}

	origURL, err := s.shortener.Extract(ctx, shortID)
	if err != nil {
		return "", userUUID, fmt.Errorf("extract short url from storage: %w", err)
	}
	return origURL, userUUID, nil
}
