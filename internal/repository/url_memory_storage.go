package repository

import (
	"context"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type MemoryURLStorage struct {
	logger  *zap.Logger
	records []model.URLStorageRecord
}

func NewMemoryURLStorage(logger *zap.Logger) *MemoryURLStorage {
	return &MemoryURLStorage{
		logger:  logger,
		records: make([]model.URLStorageRecord, 0),
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
		for _, r := range s.records {
			if r.OrigURL == url {
				return r.ShortID, nil
			}
		}
	case ShortURLType:
		for _, r := range s.records {
			if r.ShortID == url {
				return r.OrigURL, nil
			}
		}
	}
	return "", NewDataNotFoundError(nil)
}

func (s *MemoryURLStorage) Set(r *model.URLStorageRecord) error {
	s.records = append(s.records, *r)
	return nil
}

func (s *MemoryURLStorage) BatchSet(records *[]model.URLStorageRecord) error {
	s.records = append(s.records, *records...)
	return nil
}
