package service

import (
	"fmt"

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
	if err != nil {
		return "", fmt.Errorf("failed to extract short url from storage: %w", err)
	}
	return origURL, nil
}
