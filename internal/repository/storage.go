package repository

import (
	"errors"

	"go.uber.org/zap"
)

type URLStorage interface {
	Has(url string) bool
	Get(url string) (string, error)
	Set(url string, shortURL string) error
}

type MapURLStorage struct {
	storage map[string]string
	logger  *zap.Logger
}

func NewMapURLStorage(logger *zap.Logger) *MapURLStorage {
	logger = logger.With(
		zap.String("component", "storage"),
	)
	return &MapURLStorage{
		storage: make(map[string]string),
		logger:  logger,
	}
}

func (s *MapURLStorage) Has(url string) bool {
	_, ok := s.storage[url]
	s.logger.Debug("result of checking url in storage",
		zap.String("url", url),
		zap.Bool("ok", ok),
	)
	return ok
}

func (s *MapURLStorage) Get(url string) (string, error) {
	value, ok := s.storage[url]
	if !ok {
		s.logger.Debug("no data in the storage for the requested url", zap.String("url", url))
		return "", ErrURLStorageDataNotFound
	}
	s.logger.Debug("got value from storage", zap.String("url", url), zap.String("value", value))
	return value, nil
}

func (s *MapURLStorage) Set(keyURL, valueURL string) error {
	s.storage[keyURL] = valueURL
	s.logger.Debug("set value in storage",
		zap.String("keyURL", keyURL),
		zap.String("valueURL", valueURL),
	)
	return nil
}

var (
	ErrURLStorageDataNotFound = errors.New("no data in the storage for the requested url")
)
