package service

import (
	"errors"

	"github.com/alex-storchak/shortener/internal/repository"
	"go.uber.org/zap"
)

type IShortURLService interface {
	Expand(shortID string) (origURL string, err error)
}

type ShortURLService struct {
	shortener IShortener
	logger    *zap.Logger
}

func NewShortURLService(shortener IShortener, logger *zap.Logger) *ShortURLService {
	logger = logger.With(zap.String("package", "short_url_service"))
	return &ShortURLService{
		shortener: shortener,
		logger:    logger,
	}
}

func (s *ShortURLService) Expand(shortID string) (string, error) {
	origURL, err := s.shortener.Extract(shortID)
	if errors.Is(err, repository.ErrURLStorageDataNotFound) {
		s.logger.Debug("short ID not found in storage", zap.Error(err))
		return "", ErrShortURLNotFound
	} else if err != nil {
		s.logger.Debug("failed to extract url", zap.Error(err))
		return "", err
	}
	s.logger.Debug("extracted original URL", zap.String("url", origURL))
	return origURL, nil
}

var (
	ErrShortURLNotFound = errors.New("short url not found")
)
