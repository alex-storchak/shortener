package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/alex-storchak/shortener/internal/helper"
	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type IAPIShortenService interface {
	Shorten(ctx context.Context, r io.Reader) (*model.ShortenResponse, error)
}

type APIShortenService struct {
	baseURL   string
	shortener IShortener
	decoder   IJSONRequestDecoder
	logger    *zap.Logger
}

func NewAPIShortenService(bu string, s IShortener, d IJSONRequestDecoder, l *zap.Logger) *APIShortenService {
	return &APIShortenService{
		baseURL:   bu,
		shortener: s,
		decoder:   d,
		logger:    l,
	}
}

func (s *APIShortenService) Shorten(ctx context.Context, r io.Reader) (*model.ShortenResponse, error) {
	userUUID, err := helper.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user uuid from context: %w", err)
	}
	req, err := s.decoder.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request json: %w", err)
	}

	shortID, err := s.shortener.Shorten(userUUID, req.OrigURL)
	if errors.Is(err, ErrURLAlreadyExists) {
		shortURL, jpErr := url.JoinPath(s.baseURL, shortID)
		if jpErr != nil {
			return nil, fmt.Errorf("failed to build full short url path for existing url: %w", jpErr)
		}
		return &model.ShortenResponse{ShortURL: shortURL}, fmt.Errorf("tried to shorten existing url: %w", err)
	} else if err != nil {
		return nil, fmt.Errorf("failed to shorten url: %w", err)
	}

	// new url
	shortURL, err := url.JoinPath(s.baseURL, shortID)
	if err != nil {
		return nil, fmt.Errorf("failed to build full short url path for new url: %w", err)
	}
	return &model.ShortenResponse{ShortURL: shortURL}, nil
}
