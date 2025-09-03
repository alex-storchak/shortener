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
	urlStorage repository.URLStorage
	generator  IDGenerator
	logger     *zap.Logger
}

func NewShortener(
	idGenerator IDGenerator,
	urlStorage repository.URLStorage,
	logger *zap.Logger,
) *Shortener {
	logger = logger.With(
		zap.String("package", "shortener"),
	)
	return &Shortener{
		urlStorage: urlStorage,
		generator:  idGenerator,
		logger:     logger,
	}
}

func (s *Shortener) Shorten(url string) (string, error) {
	shortID, err := s.urlStorage.Get(url, repository.OrigURLType)
	if err == nil {
		s.logger.Debug("url already exists in the storage", zap.String("url", url))
		return shortID, nil
	} else if err != repository.ErrURLStorageDataNotFound {
		s.logger.Error("error retrieving url", zap.Error(err))
		return "", err
	}

	shortID, err = s.generator.Generate()
	s.logger.Debug("generated short id", zap.String("shortID", shortID))
	if err != nil {
		s.logger.Error("failed to generate short id", zap.Error(err))
		return "", ErrShortenerGenerationShortIDFailed
	}
	if err := s.urlStorage.Set(url, shortID); err != nil {
		s.logger.Error("failed to set url binding in the urlStorage", zap.Error(err))
		return "", ErrShortenerSetBindingURLStorageFailed
	}
	return shortID, nil
}

func (s *Shortener) Extract(shortID string) (string, error) {
	origURL, err := s.urlStorage.Get(shortID, repository.ShortURLType)
	if err != nil {
		return "", err
	}
	return origURL, nil
}

var (
	ErrShortenerGenerationShortIDFailed    = errors.New("failed to generate short id")
	ErrShortenerSetBindingURLStorageFailed = errors.New("failed to set url binding in the urlStorage")
)
