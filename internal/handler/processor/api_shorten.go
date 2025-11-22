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

func (s *APIShorten) Process(ctx context.Context, req model.ShortenRequest) (*model.ShortenResponse, error) {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user uuid from context: %w", err)
	}

	shortID, err := s.shortener.Shorten(userUUID, req.OrigURL)
	if errors.Is(err, service.ErrURLAlreadyExists) {
		shortURL, jpErr := url.JoinPath(s.baseURL, shortID)
		if jpErr != nil {
			return nil, fmt.Errorf("build full short url path for existing url: %w", jpErr)
		}
		return &model.ShortenResponse{ShortURL: shortURL}, fmt.Errorf("tried to shorten existing url: %w", err)
	} else if err != nil {
		return nil, fmt.Errorf("shorten url: %w", err)
	}

	// new url
	shortURL, err := url.JoinPath(s.baseURL, shortID)
	if err != nil {
		return nil, fmt.Errorf("build full short url path for new url: %w", err)
	}
	return &model.ShortenResponse{ShortURL: shortURL}, nil
}
