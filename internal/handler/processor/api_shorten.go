package processor

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

// ShortURLBuilder defines the interface for building complete short URLs from short identifiers.
type ShortURLBuilder interface {
	Build(shortID string) string
}

// APIShorten provides URL shortening functionality for JSON API requests.
// It handles the business logic for the '/api/shorten' endpoint.
type APIShorten struct {
	shortener service.URLShortener
	logger    *zap.Logger
	ub        ShortURLBuilder
}

// NewAPIShorten creates a new APIShorten processor instance.
//
// Parameters:
//   - s: URL shortener service for core shortening operations
//   - l: Structured logger for logging operations
//   - ub: URL builder for constructing complete short URLs
//
// Returns: configured APIShorten processor
func NewAPIShorten(s service.URLShortener, l *zap.Logger, ub ShortURLBuilder) *APIShorten {
	return &APIShorten{
		shortener: s,
		logger:    l,
		ub:        ub,
	}
}

// Process handles the URL shortening request for API endpoint.
// It authenticates the user, shortens the URL, and returns the response.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - req: ShortenRequest containing the original URL to shorten
//
// Returns:
//   - *ShortenResponse: response with generated short URL
//   - string: user UUID for audit purposes
//   - error: nil on success, or service error if operation fails
//
// Behavior:
//   - Returns existing short URL with ErrURLAlreadyExists if URL already exists
//   - Creates new short URL for new URLs
func (s *APIShorten) Process(ctx context.Context, req model.ShortenRequest) (*model.ShortenResponse, string, error) {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("get user uuid from context: %w", err)
	}

	shortID, err := s.shortener.Shorten(ctx, userUUID, req.OrigURL)
	if errors.Is(err, service.ErrURLAlreadyExists) {
		shortURL := s.ub.Build(shortID)
		resp := &model.ShortenResponse{ShortURL: shortURL}
		return resp, userUUID, fmt.Errorf("tried to shorten existing url: %w", err)
	} else if err != nil {
		return nil, userUUID, fmt.Errorf("shorten url: %w", err)
	}

	// new url
	shortURL := s.ub.Build(shortID)
	return &model.ShortenResponse{ShortURL: shortURL}, userUUID, nil
}
