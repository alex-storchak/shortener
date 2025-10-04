package repository

import (
	"context"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type MemoryURLStorage struct {
	logger  *zap.Logger
	records []*model.URLStorageRecord
}

func NewMemoryURLStorage(logger *zap.Logger) *MemoryURLStorage {
	return &MemoryURLStorage{
		logger:  logger,
		records: make([]*model.URLStorageRecord, 0),
	}
}

func (s *MemoryURLStorage) Close() error {
	return nil
}

func (s *MemoryURLStorage) Ping(_ context.Context) error {
	return nil
}

func (s *MemoryURLStorage) Get(url, searchByType string) (*model.URLStorageRecord, error) {
	switch searchByType {
	case OrigURLType:
		for _, r := range s.records {
			if r.OrigURL == url && !r.IsDeleted {
				return r, nil
			}
		}
	case ShortURLType:
		for _, r := range s.records {
			if r.ShortID == url {
				if r.IsDeleted {
					return nil, ErrDataDeleted
				}
				return r, nil
			}
		}
	}
	return nil, NewDataNotFoundError(nil)
}

func (s *MemoryURLStorage) Set(r *model.URLStorageRecord) error {
	s.records = append(s.records, r)
	return nil
}

func (s *MemoryURLStorage) BatchSet(records []*model.URLStorageRecord) error {
	s.records = append(s.records, records...)
	return nil
}

func (s *MemoryURLStorage) GetByUserUUID(userUUID string) ([]*model.URLStorageRecord, error) {
	var records []*model.URLStorageRecord
	for _, r := range s.records {
		if r.UserUUID == userUUID && !r.IsDeleted {
			records = append(records, r)
		}
	}
	return records, nil
}
