package repository

import (
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type MemoryUserStorage struct {
	logger *zap.Logger
	users  map[string]struct{}
}

func NewMemoryUserStorage(logger *zap.Logger) *MemoryUserStorage {
	return &MemoryUserStorage{
		logger: logger,
		users:  make(map[string]struct{}),
	}
}

func (s *MemoryUserStorage) Close() error {
	return nil
}

func (s *MemoryUserStorage) HasByUUID(uuid string) (bool, error) {
	_, ok := s.users[uuid]
	return ok, nil
}

func (s *MemoryUserStorage) Set(user *model.User) error {
	has, err := s.HasByUUID(user.UUID)
	if err != nil {
		return fmt.Errorf("failed to check if user exists before setting: %w", err)
	}
	if has {
		return fmt.Errorf("user with uuid %s already exists", user.UUID)
	}
	s.users[user.UUID] = struct{}{}
	return nil
}
