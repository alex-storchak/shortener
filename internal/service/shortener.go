package service

import (
	"errors"

	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/teris-io/shortid"
)

type Shortener struct {
	urlStorage      repository.UrlStorage
	shortUrlStorage repository.UrlStorage
	generator       *shortid.Shortid
}

func NewShortener() (*Shortener, error) {
	generator, err := shortid.New(1, shortid.DefaultABC, 1)
	if err != nil {
		return nil, errors.New("failed to instantiate short id generator")
	}
	return &Shortener{
		urlStorage:      repository.NewMapUrlStorage(),
		shortUrlStorage: repository.NewMapUrlStorage(),
		generator:       generator,
	}, nil
}

func (s Shortener) Shorten(url string) (string, error) {
	if s.urlStorage.Has(url) {
		return s.urlStorage.Get(url)
	}

	shortId, err := s.generator.Generate()
	if err != nil {
		return "", errors.New("failed to generate short url")
	}
	if err := s.urlStorage.Set(url, shortId); err != nil {
		return "", errors.New("failed to set url in the storage")
	}
	if err := s.shortUrlStorage.Set(shortId, url); err != nil {
		return "", errors.New("failed to set short url in the storage")
	}

	return shortId, nil
}

func (s Shortener) Extract(shortId string) (string, error) {
	if s.shortUrlStorage.Has(shortId) {
		return s.shortUrlStorage.Get(shortId)
	}
	return "", repository.ErrShortUrlNotFound
}
