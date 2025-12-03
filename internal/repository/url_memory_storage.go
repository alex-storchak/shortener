package repository

import (
	"context"
	"sync"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type MemoryURLStorage struct {
	logger  *zap.Logger
	records []model.URLStorageRecord
	mu      *sync.Mutex
}

func NewMemoryURLStorage(logger *zap.Logger) *MemoryURLStorage {
	return &MemoryURLStorage{
		logger:  logger,
		records: make([]model.URLStorageRecord, 0, 250000),
		mu:      &sync.Mutex{},
	}
}

func (s *MemoryURLStorage) Close() error {
	return nil
}

func (s *MemoryURLStorage) Ping(_ context.Context) error {
	return nil
}

func (s *MemoryURLStorage) Get(_ context.Context, url, searchByType string) (*model.URLStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch searchByType {
	case OrigURLType:
		for _, r := range s.records {
			if r.OrigURL == url && !r.IsDeleted {
				return &r, nil
			}
		}
	case ShortURLType:
		for _, r := range s.records {
			if r.ShortID == url {
				if r.IsDeleted {
					return nil, ErrDataDeleted
				}
				return &r, nil
			}
		}
	}
	return nil, NewDataNotFoundError(nil)
}

func (s *MemoryURLStorage) Set(_ context.Context, origURL, shortID, userUUID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, model.URLStorageRecord{
		OrigURL:  origURL,
		ShortID:  shortID,
		UserUUID: userUUID,
	})
	return nil
}

func (s *MemoryURLStorage) BatchSet(_ context.Context, records []model.URLStorageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, records...)
	return nil
}

func (s *MemoryURLStorage) GetByUserUUID(_ context.Context, userUUID string) ([]*model.URLStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := make([]*model.URLStorageRecord, 0, 50)
	for i := range s.records {
		r := &s.records[i]
		if r.UserUUID == userUUID && !r.IsDeleted {
			records = append(records, r)
		}
	}
	return records, nil
}

func (s *MemoryURLStorage) DeleteBatch(_ context.Context, urls model.URLDeleteBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ProcessMemDeleteBatch(s.records, urls)
	return nil
}

func ProcessMemDeleteBatch(records []model.URLStorageRecord, urls model.URLDeleteBatch) {
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

	for i := range records {
		r := &records[i]
		if userMap, exists := deleteMap[r.ShortID]; exists {
			if userMap[r.UserUUID] {
				r.IsDeleted = true
			}
		}
	}
}
