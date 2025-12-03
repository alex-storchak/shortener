package processor

import (
	"context"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type ShortURLBuilder interface {
	Build(shortID string) string
}

type APIShorten struct {
	shortener service.URLShortener
	logger    *zap.Logger
	ub        ShortURLBuilder
}

func NewAPIShorten(s service.URLShortener, l *zap.Logger, ub ShortURLBuilder) *APIShorten {
	return &APIShorten{
		shortener: s,
		logger:    l,
		ub:        ub,
	}
}

func (s *APIShorten) Process(ctx context.Context, req model.ShortenRequest) (*model.ShortenResponse, string, error) {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("get user uuid from context: %w", err)
	}

	shortID, err := s.shortener.Shorten(ctx, userUUID, req.OrigURL)
	if errors.Is(err, service.ErrURLAlreadyExists) {
		shortURL := s.ub.Build(shortID)
		resp := &model.ShortenResponse{ShortURL: shortURL}
		return resp, userUUID, fmt.Errorf("tried to shorten existing url: %w", err)
	} else if err != nil {
		return nil, userUUID, fmt.Errorf("shorten url: %w", err)
	}

	// new url
	shortURL := s.ub.Build(shortID)
	return &model.ShortenResponse{ShortURL: shortURL}, userUUID, nil
}
