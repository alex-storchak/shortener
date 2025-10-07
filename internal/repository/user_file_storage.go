package repository

import (
	"fmt"
	"sync"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type FileUserStorage struct {
	logger   *zap.Logger
	fileMgr  *FileManager
	fileScnr *URLFileScanner
	users    map[string]struct{}
	mu       *sync.Mutex
}

func NewFileUserStorage(
	logger *zap.Logger,
	fm *FileManager,
	fs *URLFileScanner,
) (*FileUserStorage, error) {
	storage := &FileUserStorage{
		logger:   logger,
		fileMgr:  fm,
		fileScnr: fs,
		mu:       &sync.Mutex{},
	}

	if err := storage.restoreFromFile(); err != nil {
		return nil, fmt.Errorf("failed to restore storage from file: %w", err)
	}
	return storage, nil
}

func (s *FileUserStorage) Close() error {
	return s.fileMgr.close()
}

func (s *FileUserStorage) HasByUUID(uuid string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hasByUUIDUnsafe(uuid)
}

func (s *FileUserStorage) hasByUUIDUnsafe(uuid string) (bool, error) {
	_, ok := s.users[uuid]
	return ok, nil
}

func (s *FileUserStorage) Set(user *model.User) error {
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

func (s *FileUserStorage) restoreFromFile() error {
	file, err := s.fileMgr.openForAppend(false)
	if err != nil {
		return fmt.Errorf("failed to open requested file: %w", err)
	}

	records, err := s.fileScnr.scan(file)
	if err != nil {
		if cErr := s.fileMgr.close(); cErr != nil {
			return fmt.Errorf("failed to close requested file: %w", cErr)
		}
		return fmt.Errorf("failed to scan data from requested file: %w", err)
	}

	users := make(map[string]struct{})
	for _, record := range records {
		users[record.UserUUID] = struct{}{}
	}
	s.users = users
	return nil
}
