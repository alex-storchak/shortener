package service

import (
	"errors"
	"fmt"
	"net/url"

	"go.uber.org/zap"
)

type IMainPageService interface {
	Shorten(body []byte) (shortURL string, err error)
}

type MainPageService struct {
	baseURL   string
	shortener IShortener
	logger    *zap.Logger
}

func NewMainPageService(bu string, s IShortener, l *zap.Logger) *MainPageService {
	return &MainPageService{
		baseURL:   bu,
		shortener: s,
		logger:    l,
	}
}

func (s *MainPageService) Shorten(body []byte) (string, error) {
	origURL := string(body)
	shortID, err := s.shortener.Shorten(origURL)
	if errors.Is(err, ErrURLAlreadyExists) {
		shortURL, jpErr := url.JoinPath(s.baseURL, shortID)
		if jpErr != nil {
			return "", fmt.Errorf("failed to build full short url path for existing url: %w", jpErr)
		}
		return shortURL, fmt.Errorf("tried to shorten existing url: %w", err)
	} else if err != nil {
		return "", fmt.Errorf("failed to shorten url: %w", err)
	}

	// new url
	shortURL, err := url.JoinPath(s.baseURL, shortID)
	if err != nil {
		return "", fmt.Errorf("failed to build full short url path for new url: %w", err)
	}
	return shortURL, nil
}
