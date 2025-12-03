package processor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

// APIShortenBatch provides batch URL shortening functionality for multiple URLs.
// It handles the business logic for the '/api/shorten/batch' endpoint.
type APIShortenBatch struct {
	shortener service.URLShortener
	logger    *zap.Logger
	ub        ShortURLBuilder
}

// NewAPIShortenBatch creates a new APIShortenBatch processor instance.
//
// Parameters:
//   - s: URL shortener service for batch shortening operations
//   - l: Structured logger for logging operations
//   - ub: URL builder for constructing complete short URLs
//
// Returns: configured APIShortenBatch processor
func NewAPIShortenBatch(s service.URLShortener, l *zap.Logger, ub ShortURLBuilder) *APIShortenBatch {
	return &APIShortenBatch{
		shortener: s,
		logger:    l,
		ub:        ub,
	}
}

// Process handles batch URL shortening requests for multiple URLs.
// It processes all URLs in a single operation while maintaining request-response correlation.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - items: model.BatchShortenRequest containing URLs with correlation IDs
//
// Returns:
//   - model.BatchShortenResponse: response with shortened URLs and correlation IDs
//   - error: nil on success, or service error if operation fails
func (s *APIShortenBatch) Process(ctx context.Context, items model.BatchShortenRequest) (model.BatchShortenResponse, error) {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user uuid from context: %w", err)
	}

	origURLs := s.buildURLList(items)
	shortIDs, err := s.shortener.ShortenBatch(ctx, userUUID, origURLs)
	if err != nil {
		return nil, fmt.Errorf("shorten batch: %w", err)
	}

	resp, err := s.buildResponse(items, shortIDs)
	if err != nil {
		return nil, fmt.Errorf("build response: %w", err)
	}

	return resp, nil
}

// buildURLList creates a list of original URLs from the batch request.
//
// Parameters:
//   - reqItems: model.BatchShortenRequest containing URLs with correlation IDs
//
// Returns:
//   - []string: list of original URLs
func (s *APIShortenBatch) buildURLList(reqItems model.BatchShortenRequest) []string {
	origURLs := make([]string, len(reqItems))
	for i, item := range reqItems {
		origURLs[i] = item.OriginalURL
	}
	return origURLs
}

// buildResponse creates a batch response with shortened URLs and correlation IDs.
//
// Parameters:
//   - reqItems: model.BatchShortenRequest containing URLs with correlation IDs
//   - shortIDs: list of short IDs generated for the original URLs
//
// Returns:
//   - model.BatchShortenResponse: list of response items with correlation IDs and short URLs
//   - error: nil on success, or error if response construction fails
func (s *APIShortenBatch) buildResponse(
	reqItems model.BatchShortenRequest,
	shortIDs []string,
) (model.BatchShortenResponse, error) {
	resp := make(model.BatchShortenResponse, len(reqItems))
	for i, item := range reqItems {
		shortURL := s.ub.Build(shortIDs[i])
		resp[i] = model.BatchShortenResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		}
	}
	return resp, nil
}
