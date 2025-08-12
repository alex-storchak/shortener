package service

import (
	"errors"

	"github.com/alex-storchak/shortener/internal/repository"
)

type IShortener interface {
	Shorten(url string) (string, error)
	Extract(shortID string) (string, error)
}

type Shortener struct {
	urlStorage      repository.URLStorage
	shortURLStorage repository.URLStorage
	generator       IDGenerator
}

func NewShortener(
	idGenerator IDGenerator,
	urlStorage repository.URLStorage,
	shortURLStorage repository.URLStorage,
) *Shortener {
	return &Shortener{
		urlStorage:      urlStorage,
		shortURLStorage: shortURLStorage,
		generator:       idGenerator,
	}
}

func (s Shortener) Shorten(url string) (string, error) {
	if s.urlStorage.Has(url) {
		return s.urlStorage.Get(url)
	}

	shortID, err := s.generator.Generate()
	if err != nil {
		return "", ErrShortenerGenerationShortIDFailed
	}
	if err := s.urlStorage.Set(url, shortID); err != nil {
		return "", ErrShortenerSetBindingURLStorageFailed
	}
	if err := s.shortURLStorage.Set(shortID, url); err != nil {
		return "", ErrShortenerSetBindingShortURLStorageFailed
	}

	return shortID, nil
}

func (s Shortener) Extract(shortID string) (string, error) {
	if s.shortURLStorage.Has(shortID) {
		return s.shortURLStorage.Get(shortID)
	}
	return "", ErrShortenerShortIDNotFound
}

var (
	ErrShortenerGenerationShortIDFailed         = errors.New("failed to generate short id")
	ErrShortenerSetBindingURLStorageFailed      = errors.New("failed to set url binding in the urlStorage")
	ErrShortenerSetBindingShortURLStorageFailed = errors.New("failed to set url binding in the shortURLStorage")
	ErrShortenerShortIDNotFound                 = errors.New("short id binding not found in the shortURLStorage")
)
