package repository

import (
	"context"

	"go.uber.org/zap"
)

type MemoryURLStorage struct {
	logger      *zap.Logger
	origToShort map[string]string
	shortToOrig map[string]string
}

func NewMemoryURLStorage(logger *zap.Logger) *MemoryURLStorage {
	return &MemoryURLStorage{
		logger:      logger,
		origToShort: make(map[string]string),
		shortToOrig: make(map[string]string),
	}
}

func (s *MemoryURLStorage) Close() error {
	return nil
}

func (s *MemoryURLStorage) Ping(_ context.Context) error {
	return nil
}

func (s *MemoryURLStorage) Get(url, searchByType string) (string, error) {
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
	return "", NewDataNotFoundError(nil)
}

func (s *MemoryURLStorage) Set(origURL, shortURL string) error {
	s.origToShort[origURL] = shortURL
	s.shortToOrig[shortURL] = origURL
	return nil
}

func (s *MemoryURLStorage) BatchSet(binds *[]URLBind) error {
	for _, b := range *binds {
		s.origToShort[b.OrigURL] = b.ShortID
		s.shortToOrig[b.ShortID] = b.OrigURL
	}
	return nil
}
