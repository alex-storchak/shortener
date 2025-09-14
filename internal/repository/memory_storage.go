package repository

import (
	"go.uber.org/zap"
)

type MemoryURLStorage struct {
	logger      *zap.Logger
	origToShort map[string]string
	shortToOrig map[string]string
}

func NewMemoryURLStorage(logger *zap.Logger) *MemoryURLStorage {
	logger = logger.With(zap.String("component", "memory_storage"))
	return &MemoryURLStorage{
		logger:      logger,
		origToShort: make(map[string]string),
		shortToOrig: make(map[string]string),
	}
}

func (s *MemoryURLStorage) Get(url, searchByType string) (string, error) {
	s.logger.Debug("Getting url from memory storage",
		zap.String("url", url),
		zap.String("searchByType", searchByType),
	)
	switch searchByType {
	case OrigURLType:
		if v, ok := s.origToShort[url]; ok {
			return v, nil
		}
	case ShortURLType:
		if v, ok := s.shortToOrig[url]; ok {
			return v, nil
		}
	}
	return "", ErrURLStorageDataNotFound
}

func (s *MemoryURLStorage) Set(origURL, shortURL string) error {
	s.logger.Debug("Setting url binding in memory storage",
		zap.String("origURL", origURL),
		zap.String("shortURL", shortURL),
	)
	s.origToShort[origURL] = shortURL
	s.shortToOrig[shortURL] = origURL
	return nil
}
