package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type IShortURLService interface {
	Expand(ctx context.Context, shortID string) (origURL string, err error)
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

func (s *ShortURLService) Expand(ctx context.Context, shortID string) (string, error) {
	/*	userUUID, err := helper.GetCtxUserUUID(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get user uuid from context: %w", err)
		}*/
	origURL, err := s.shortener.Extract( /*userUUID, */ shortID)
	if err != nil {
		return "", fmt.Errorf("failed to extract short url from storage: %w", err)
	}
	return origURL, nil
}
