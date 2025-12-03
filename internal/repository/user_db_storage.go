package repository

import (
	"context"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type UserDBManager interface {
	HasByUUID(ctx context.Context, uuid string) (bool, error)
	Persist(ctx context.Context, user *model.User) error
	Close() error
}

type DBUserStorage struct {
	logger *zap.Logger
	dbMgr  UserDBManager
}

func NewDBUserStorage(logger *zap.Logger, dbm UserDBManager) *DBUserStorage {
	return &DBUserStorage{
		logger: logger,
		dbMgr:  dbm,
	}
}

func (s *DBUserStorage) Close() error {
	return s.dbMgr.Close()
}

func (s *DBUserStorage) HasByUUID(uuid string) (bool, error) {
	has, err := s.dbMgr.HasByUUID(context.Background(), uuid)
	if err != nil {
		return false, fmt.Errorf("check if user exists: %w", err)
	}
	return has, nil
}

func (s *DBUserStorage) Set(user *model.User) error {
	if err := s.dbMgr.Persist(context.Background(), user); err != nil {
		return fmt.Errorf("persist user: %w", err)
	}
	return nil
}
