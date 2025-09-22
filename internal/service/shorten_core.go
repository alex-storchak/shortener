package service

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

type IShortenCore interface {
	Shorten(origURL string) (ShortURL string, shortID string, err error)
}

type ShortenCore struct {
	shortener IShortener
	baseURL   string
	logger    *zap.Logger
}

func NewShortenCore(shortener IShortener, baseURL string, logger *zap.Logger) *ShortenCore {
	return &ShortenCore{
		shortener: shortener,
		baseURL:   baseURL,
		logger:    logger,
	}
}

func (c *ShortenCore) Shorten(origURL string) (string, string, error) {
	if len(origURL) == 0 {
		return "", "", ErrEmptyInputURL
	}

	shortID, err := c.shortener.Shorten(origURL)
	if err != nil {
		if errors.Is(err, ErrURLAlreadyExists) {
			shortURL := fmt.Sprintf("%s/%s", c.baseURL, shortID)
			return shortURL, shortID, ErrURLAlreadyExists
		}
		return "", "", err
	}
	shortURL := fmt.Sprintf("%s/%s", c.baseURL, shortID)
	return shortURL, shortID, nil
}

var (
	ErrEmptyInputURL = errors.New("empty url input")
)
