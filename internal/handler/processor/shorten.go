package processor

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/service"
)

type Shorten struct {
	shortener service.URLShortener
	logger    *zap.Logger
	ub        ShortURLBuilder
}

func NewShorten(s service.URLShortener, l *zap.Logger, ub ShortURLBuilder) *Shorten {
	return &Shorten{
		shortener: s,
		logger:    l,
		ub:        ub,
	}
}

func (s *Shorten) Process(ctx context.Context, body []byte) (string, string, error) {
	origURL := string(body)
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return "", "", fmt.Errorf("get user uuid from context: %w", err)
	}
	shortID, err := s.shortener.Shorten(ctx, userUUID, origURL)
	if errors.Is(err, service.ErrURLAlreadyExists) {
		shortURL := s.ub.Build(shortID)
		return shortURL, userUUID, fmt.Errorf("tried to shorten existing url: %w", err)
	} else if err != nil {
		return "", userUUID, fmt.Errorf("shorten url: %w", err)
	}

	// new url
	shortURL := s.ub.Build(shortID)
	return shortURL, userUUID, nil
}
