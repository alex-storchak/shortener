package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
	repo "github.com/alex-storchak/shortener/internal/repository"
)

// IDGenerator defines the interface for generating unique short identifiers.
type IDGenerator interface {
	Generate() (string, error)
}

// Pinger defines the interface for checking service readiness.
type Pinger interface {
	IsReady(ctx context.Context) error
}

// URLShortener defines the main interface for URL shortening operations.
// It provides methods for shortening URLs, extracting original URLs,
// batch operations, and user-specific URL management.
type URLShortener interface {
	Shorten(ctx context.Context, userUUID string, url string) (shortID string, err error)
	Extract(ctx context.Context, shortID string) (OrigURL string, err error)
	ShortenBatch(ctx context.Context, userUUID string, urls []string) ([]string, error)
	GetUserURLs(ctx context.Context, userUUID string) ([]*model.URLStorageRecord, error)
	DeleteBatch(ctx context.Context, urls model.URLDeleteBatch) error
	Count(ctx context.Context) (int, error)
}

// PingableURLShortener combines URL shortening functionality with health checking capability.
type PingableURLShortener interface {
	URLShortener
	Pinger
}

// Shortener implements the URLShortener interface and provides URL shortening services.
// It uses a storage backend for persistence and an ID generator for creating short identifiers.
type Shortener struct {
	urlStorage repo.URLStorage
	generator  IDGenerator
	logger     *zap.Logger
}

// NewShortener creates a new instance of Shortener with the specified dependencies.
//
// Parameters:
//   - idGenerator: generator for creating unique short IDs
//   - urlStorage: storage backend for URL persistence
//   - logger: structured logger for logging operations
//
// Returns:
//   - *Shortener: configured Shortener instance
func NewShortener(
	idGenerator IDGenerator,
	urlStorage repo.URLStorage,
	logger *zap.Logger,
) *Shortener {
	return &Shortener{
		urlStorage: urlStorage,
		generator:  idGenerator,
		logger:     logger,
	}
}

// Shorten creates a short URL for the provided original URL and associates it with a user.
// If the URL already exists in storage, it returns the existing short ID with ErrURLAlreadyExists.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userUUID: unique identifier of the user making the request
//   - url: original URL to be shortened
//
// Returns:
//   - string: generated short identifier
//   - error: nil on success, or one of the service errors
//
// Errors:
//   - ErrEmptyInputURL: when provided URL is empty
//   - ErrURLAlreadyExists: when URL already has a short identifier in storage
func (s *Shortener) Shorten(ctx context.Context, userUUID string, url string) (string, error) {
	if len(url) == 0 {
		return "", ErrEmptyInputURL
	}

	var (
		r   *model.URLStorageRecord
		err error
	)
	r, err = s.urlStorage.Get(ctx, url, repo.OrigURLType)
	if err == nil {
		return r.ShortID, ErrURLAlreadyExists
	}
	var nfErr *repo.DataNotFoundError
	if !errors.As(err, &nfErr) {
		return "", fmt.Errorf("retrieve url from storage: %w", err)
	}

	shortID, err := s.generator.Generate()
	if err != nil {
		return "", fmt.Errorf("generate short id: %w", err)
	}
	if err := s.urlStorage.Set(ctx, url, shortID, userUUID); err != nil {
		return "", fmt.Errorf("set url binding in storage: %w", err)
	}
	return shortID, nil
}

// Extract retrieves the original URL for a given short identifier.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - shortID: short identifier to look up
//
// Returns:
//   - string: original URL associated with the short identifier
//   - error: nil on success, or storage error if URL not found or deleted
func (s *Shortener) Extract(ctx context.Context, shortID string) (string, error) {
	r, err := s.urlStorage.Get(ctx, shortID, repo.ShortURLType)
	if err != nil {
		return "", fmt.Errorf("retrieve short url from storage: %w", err)
	}
	return r.OrigURL, nil
}

// ShortenBatch creates short URLs for multiple original URLs in a single operation.
// It efficiently handles existing URLs by reusing their short identifiers.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userUUID: unique identifier of the user making the request
//   - urls: slice of original URLs to be shortened
//
// Returns:
//   - []string: slice of short identifiers corresponding to input URLs
//   - error: nil on success, or service error if operation fails
//
// Errors:
//   - ErrEmptyInputBatch: when provided URL slice is empty
//   - ErrEmptyInputURL: when any URL in the batch is empty
func (s *Shortener) ShortenBatch(ctx context.Context, userUUID string, urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, ErrEmptyInputBatch
	}

	res, toPersist, err := s.segregateBatch(ctx, userUUID, urls)
	if err != nil {
		return nil, fmt.Errorf("segregate batch: %w", err)
	}
	if len(toPersist) > 0 {
		if err := s.urlStorage.BatchSet(ctx, toPersist); err != nil {
			return nil, fmt.Errorf("set url bindings batch in storage: %w", err)
		}
	}
	return res, nil
}

// segregateBatch processes a batch of URLs, separating existing URLs from new ones.
// It returns short identifiers for all URLs and records that need to be persisted.
func (s *Shortener) segregateBatch(ctx context.Context, userUUID string, urls []string) ([]string, []model.URLStorageRecord, error) {
	res := make([]string, len(urls))
	toPersist := make([]model.URLStorageRecord, 0)

	for i, u := range urls {
		if u == "" {
			return nil, nil, ErrEmptyInputURL
		}

		r, err := s.urlStorage.Get(ctx, u, repo.OrigURLType)
		if err == nil {
			res[i] = r.ShortID
			continue
		}
		var nfErr *repo.DataNotFoundError
		if !errors.As(err, &nfErr) {
			return nil, nil, fmt.Errorf("retrieve url from storage: %w", err)
		}

		urlBindItem, err := s.prepareURLBindToPersistItem(userUUID, u)
		if err != nil {
			return nil, nil, fmt.Errorf("prepare url bind to persist item: %w", err)
		}
		toPersist = append(toPersist, urlBindItem)
		res[i] = urlBindItem.ShortID
	}
	return res, toPersist, nil
}

// prepareURLBindToPersistItem creates a URLStorageRecord with a generated short ID.
func (s *Shortener) prepareURLBindToPersistItem(userUUID string, origURL string) (model.URLStorageRecord, error) {
	shortID, err := s.generator.Generate()
	if err != nil {
		return model.URLStorageRecord{}, fmt.Errorf("batch. generate short id: %w", err)
	}
	return model.URLStorageRecord{OrigURL: origURL, ShortID: shortID, UserUUID: userUUID}, nil
}

// IsReady checks if the service is ready to handle requests by pinging the storage backend.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//
// Returns:
//   - error: nil if storage is accessible, or error if connection fails
func (s *Shortener) IsReady(ctx context.Context) error {
	return s.urlStorage.Ping(ctx)
}

// GetUserURLs retrieves all URLs shortened by a specific user.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userUUID: unique identifier of the user
//
// Returns:
//   - []*model.URLStorageRecord: slice of URL records belonging to the user
//   - error: nil on success, or storage error if operation fails
func (s *Shortener) GetUserURLs(ctx context.Context, userUUID string) ([]*model.URLStorageRecord, error) {
	return s.urlStorage.GetByUserUUID(ctx, userUUID)
}

// DeleteBatch marks multiple URLs as deleted in a batch operation.
// Only URLs belonging to the specified user can be deleted.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - urls: batch of URLs to delete with user identifiers
//
// Returns:
//   - error: nil on success, or storage error if operation fails
func (s *Shortener) DeleteBatch(ctx context.Context, urls model.URLDeleteBatch) error {
	return s.urlStorage.DeleteBatch(ctx, urls)
}

// Count counts the amount of shortened URLs in storage.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//
// Returns:
//   - int: amount of shortened URLs in storage
//   - error: nil on success, or storage error if operation fails
func (s *Shortener) Count(ctx context.Context) (int, error) {
	return s.urlStorage.Count(ctx)
}

// Common service errors
var (
	// ErrURLAlreadyExists is returned when attempting to shorten a URL that already exists in storage.
	ErrURLAlreadyExists = errors.New("url already exists")

	// ErrEmptyInputURL is returned when an empty URL is provided for shortening.
	ErrEmptyInputURL = errors.New("empty url in the input")

	// ErrEmptyInputBatch is returned when an empty batch is provided for batch operations.
	ErrEmptyInputBatch = errors.New("empty batch provided")
)
