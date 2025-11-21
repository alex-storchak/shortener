package service

import (
	"fmt"

	"go.uber.org/zap"
)

type ShortURLService struct {
	shortener URLShortener
	logger    *zap.Logger
}

func NewShortURLService(shortener URLShortener, logger *zap.Logger) *ShortURLService {
	return &ShortURLService{
		shortener: shortener,
		logger:    logger,
	}
}

func (s *ShortURLService) Process(shortID string) (string, error) {
	origURL, err := s.shortener.Extract(shortID)
	if err != nil {
		return "", fmt.Errorf("extract short url from storage: %w", err)
	}
	return origURL, nil
}
