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
	logger = logger.With(zap.String("package", "shorten_core"))
	return &ShortenCore{
		shortener: shortener,
		baseURL:   baseURL,
		logger:    logger,
	}
}

func (c *ShortenCore) Shorten(origURL string) (string, string, error) {
	if len(origURL) == 0 {
		c.logger.Debug("empty input for shorten")
		return "", "", ErrEmptyInputURL
	}

	shortID, err := c.shortener.Shorten(origURL)
	if err != nil {
		c.logger.Debug("failed to shorten url", zap.Error(err))
		return "", "", err
	}
	shortURL := fmt.Sprintf("%s/%s", c.baseURL, shortID)
	c.logger.Debug("shortened url", zap.String("url", shortURL))
	return shortURL, shortID, nil
}

var (
	ErrEmptyInputURL = errors.New("empty url input")
)
