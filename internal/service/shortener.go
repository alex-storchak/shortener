package service

import (
	"errors"

	"github.com/alex-storchak/shortener/internal/repository"
	"go.uber.org/zap"
)

type IShortener interface {
	Shorten(url string) (string, error)
	Extract(shortID string) (string, error)
}

type Shortener struct {
	urlStorage      repository.URLStorage
	shortURLStorage repository.URLStorage
	generator       IDGenerator
	logger          *zap.Logger
}

func NewShortener(
	idGenerator IDGenerator,
	urlStorage repository.URLStorage,
	shortURLStorage repository.URLStorage,
	logger *zap.Logger,
) *Shortener {
	logger = logger.With(
		zap.String("package", "shortener"),
	)
	return &Shortener{
		urlStorage:      urlStorage,
		shortURLStorage: shortURLStorage,
		generator:       idGenerator,
		logger:          logger,
	}
}

func (s *Shortener) Shorten(url string) (string, error) {
	if s.urlStorage.Has(url) {
		s.logger.Debug("url already exists in the storage", zap.String("url", url))
		return s.urlStorage.Get(url)
	}

	shortID, err := s.generator.Generate()
	s.logger.Debug("generated short id", zap.String("shortID", shortID))
	if err != nil {
		s.logger.Error("failed to generate short id", zap.Error(err))
		return "", ErrShortenerGenerationShortIDFailed
	}
	if err := s.urlStorage.Set(url, shortID); err != nil {
		s.logger.Error("failed to set url binding in the urlStorage", zap.Error(err))
		return "", ErrShortenerSetBindingURLStorageFailed
	}
	if err := s.shortURLStorage.Set(shortID, url); err != nil {
		s.logger.Error("failed to set url binding in the shortURLStorage", zap.Error(err))
		return "", ErrShortenerSetBindingShortURLStorageFailed
	}

	return shortID, nil
}

func (s *Shortener) Extract(shortID string) (string, error) {
	if s.shortURLStorage.Has(shortID) {
		s.logger.Debug("short id already exists in the storage")
		return s.shortURLStorage.Get(shortID)
	}
	s.logger.Debug("short id not found in the storage")
	return "", ErrShortenerShortIDNotFound
}

var (
	ErrShortenerGenerationShortIDFailed         = errors.New("failed to generate short id")
	ErrShortenerSetBindingURLStorageFailed      = errors.New("failed to set url binding in the urlStorage")
	ErrShortenerSetBindingShortURLStorageFailed = errors.New("failed to set url binding in the shortURLStorage")
	ErrShortenerShortIDNotFound                 = errors.New("short id binding not found in the shortURLStorage")
)
