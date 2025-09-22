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
	return &ShortURLService{
		shortener: shortener,
		logger:    logger,
	}
}

func (s *ShortURLService) Expand(shortID string) (string, error) {
	origURL, err := s.shortener.Extract(shortID)
	if errors.Is(err, repository.ErrURLStorageDataNotFound) {
		return "", ErrShortURLNotFound
	} else if err != nil {
		return "", err
	}
	return origURL, nil
}

var (
	ErrShortURLNotFound = errors.New("short url not found")
)
