package processor

import (
	"fmt"

	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
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

func (s *Expand) Process(shortID string) (string, error) {
	origURL, err := s.shortener.Extract(shortID)
	if err != nil {
		return "", fmt.Errorf("extract short url from storage: %w", err)
	}
	return origURL, nil
}
