package processor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/service"
)

type Expand struct {
	shortener service.URLShortener
	logger    *zap.Logger
}

func NewExpand(shortener service.URLShortener, logger *zap.Logger) *Expand {
	return &Expand{
		shortener: shortener,
		logger:    logger,
	}
}

func (s *Expand) Process(ctx context.Context, shortID string) (string, string, error) {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		s.logger.Debug("failed to get user uuid from context", zap.Error(err))
		userUUID = ""
	}

	origURL, err := s.shortener.Extract(ctx, shortID)
	if err != nil {
		return "", userUUID, fmt.Errorf("extract short url from storage: %w", err)
	}
	return origURL, userUUID, nil
}
