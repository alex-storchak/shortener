package processor

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type Shorten struct {
	baseURL   string
	shortener service.URLShortener
	logger    *zap.Logger
}

func NewShorten(bu string, s service.URLShortener, l *zap.Logger) *Shorten {
	return &Shorten{
		baseURL:   bu,
		shortener: s,
		logger:    l,
	}
}

func (s *Shorten) Process(ctx context.Context, body []byte) (string, error) {
	origURL := string(body)
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return "", fmt.Errorf("get user uuid from context: %w", err)
	}
	shortID, err := s.shortener.Shorten(userUUID, origURL)
	if errors.Is(err, service.ErrURLAlreadyExists) {
		shortURL, jpErr := url.JoinPath(s.baseURL, shortID)
		if jpErr != nil {
			return "", fmt.Errorf("build full short url path for existing url: %w", jpErr)
		}
		return shortURL, fmt.Errorf("tried to shorten existing url: %w", err)
	} else if err != nil {
		return "", fmt.Errorf("shorten url: %w", err)
	}

	// new url
	shortURL, err := url.JoinPath(s.baseURL, shortID)
	if err != nil {
		return "", fmt.Errorf("build full short url path for new url: %w", err)
	}
	return shortURL, nil
}
