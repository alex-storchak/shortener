package processor

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type APIShorten struct {
	baseURL   string
	shortener service.URLShortener
	logger    *zap.Logger
}

func NewAPIShorten(bu string, s service.URLShortener, l *zap.Logger) *APIShorten {
	return &APIShorten{
		baseURL:   bu,
		shortener: s,
		logger:    l,
	}
}

func (s *APIShorten) Process(ctx context.Context, req model.ShortenRequest) (*model.ShortenResponse, string, error) {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("get user uuid from context: %w", err)
	}

	shortID, err := s.shortener.Shorten(ctx, userUUID, req.OrigURL)
	if errors.Is(err, service.ErrURLAlreadyExists) {
		shortURL, jpErr := url.JoinPath(s.baseURL, shortID)
		if jpErr != nil {
			return nil, userUUID, fmt.Errorf("build full short url path for existing url: %w", jpErr)
		}
		resp := &model.ShortenResponse{ShortURL: shortURL}
		return resp, userUUID, fmt.Errorf("tried to shorten existing url: %w", err)
	} else if err != nil {
		return nil, userUUID, fmt.Errorf("shorten url: %w", err)
	}

	// new url
	shortURL, err := url.JoinPath(s.baseURL, shortID)
	if err != nil {
		return nil, userUUID, fmt.Errorf("build full short url path for new url: %w", err)
	}
	return &model.ShortenResponse{ShortURL: shortURL}, userUUID, nil
}
