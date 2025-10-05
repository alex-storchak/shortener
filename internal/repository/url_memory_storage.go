package repository

import (
	"context"
	"sync"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type MemoryURLStorage struct {
	logger  *zap.Logger
	records []*model.URLStorageRecord
	mu      sync.RWMutex
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
	s.mu.RLock()
	defer s.mu.RUnlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, r)
	return nil
}

func (s *MemoryURLStorage) BatchSet(records []*model.URLStorageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, records...)
	return nil
}

func (s *MemoryURLStorage) GetByUserUUID(userUUID string) ([]*model.URLStorageRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var records []*model.URLStorageRecord
	for _, r := range s.records {
		if r.UserUUID == userUUID && !r.IsDeleted {
			records = append(records, r)
		}
	}
	return records, nil
}

func (s *MemoryURLStorage) DeleteBatch(urls model.URLDeleteBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ProcessMemDeleteBatch(s.records, urls)
	return nil
}

func ProcessMemDeleteBatch(records []*model.URLStorageRecord, urls model.URLDeleteBatch) {
	if len(urls) == 0 {
		return
	}

	deleteMap := make(map[string]map[string]bool)
	for _, u := range urls {
		if deleteMap[u.ShortID] == nil {
			deleteMap[u.ShortID] = make(map[string]bool)
		}
		deleteMap[u.ShortID][u.UserUUID] = true
	}

	for _, r := range records {
		if userMap, exists := deleteMap[r.ShortID]; exists {
			if userMap[r.UserUUID] {
				r.IsDeleted = true
			}
		}
	}
}
