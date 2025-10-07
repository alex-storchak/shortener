package repository

import (
	"fmt"
	"sync"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type MemoryUserStorage struct {
	logger *zap.Logger
	users  map[string]struct{}
	mu     *sync.Mutex
}

func NewMemoryUserStorage(logger *zap.Logger) *MemoryUserStorage {
	return &MemoryUserStorage{
		logger: logger,
		users:  make(map[string]struct{}),
		mu:     &sync.Mutex{},
	}
}

func (s *MemoryUserStorage) Close() error {
	return nil
}

func (s *MemoryUserStorage) HasByUUID(uuid string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hasByUUIDUnsafe(uuid)
}

func (s *MemoryUserStorage) hasByUUIDUnsafe(uuid string) (bool, error) {
	_, ok := s.users[uuid]
	return ok, nil
}

func (s *MemoryUserStorage) Set(user *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	has, err := s.hasByUUIDUnsafe(user.UUID)
	if err != nil {
		return fmt.Errorf("failed to check if user exists before setting: %w", err)
	}
	if has {
		return fmt.Errorf("user with uuid %s already exists", user.UUID)
	}
	s.users[user.UUID] = struct{}{}
	return nil
}
