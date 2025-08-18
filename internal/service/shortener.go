package service

import (
	"errors"

	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/teris-io/shortid"
)

type Shortener struct {
	urlStorage      repository.URLStorage
	shortURLStorage repository.URLStorage
	generator       *shortid.Shortid
}

func NewShortener() (*Shortener, error) {
	generator, err := shortid.New(1, shortid.DefaultABC, 1)
	if err != nil {
		return nil, errors.New("failed to instantiate short id generator")
	}
	return &Shortener{
		urlStorage:      repository.NewMapURLStorage(),
		shortURLStorage: repository.NewMapURLStorage(),
		generator:       generator,
	}, nil
}

func (s Shortener) Shorten(url string) (string, error) {
	if s.urlStorage.Has(url) {
		return s.urlStorage.Get(url)
	}

	shortID, err := s.generator.Generate()
	if err != nil {
		return "", errors.New("failed to generate short url")
	}
	if err := s.urlStorage.Set(url, shortID); err != nil {
		return "", errors.New("failed to set url in the storage")
	}
	if err := s.shortURLStorage.Set(shortID, url); err != nil {
		return "", errors.New("failed to set short url in the storage")
	}

	return shortID, nil
}

func (s Shortener) Extract(shortID string) (string, error) {
	if s.shortURLStorage.Has(shortID) {
		return s.shortURLStorage.Get(shortID)
	}
	return "", repository.ErrShortURLNotFound
}
