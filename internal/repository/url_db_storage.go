package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type IURLDBManager interface {
	GetByOriginalURL(ctx context.Context, origURL string /*, userUUID string*/) (string, error)
	GetByShortID(ctx context.Context, shortID string /*, userUUID string*/) (string, error)
	Persist(ctx context.Context, r *model.URLStorageRecord) error
	PersistBatch(ctx context.Context, binds *[]model.URLStorageRecord) error
	Ping(ctx context.Context) error
	Close() error
	GetByUserUUID(ctx context.Context, userUUID string) (*[]model.URLStorageRecord, error)
}

type DBURLStorage struct {
	logger *zap.Logger
	dbMgr  IURLDBManager
}

func NewDBURLStorage(logger *zap.Logger, dbm IURLDBManager) *DBURLStorage {
	return &DBURLStorage{
		logger: logger,
		dbMgr:  dbm,
	}
}

func (s *DBURLStorage) Close() error {
	return s.dbMgr.Close()
}

func (s *DBURLStorage) Ping(ctx context.Context) error {
	if err := s.dbMgr.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}
	return nil
}

func (s *DBURLStorage) Get( /*userUUID string, */ url, searchByType string) (string, error) {
	var err error
	if searchByType == OrigURLType {
		url, err = s.dbMgr.GetByOriginalURL(context.Background(), url /*, userUUID*/)
	} else if searchByType == ShortURLType {
		url, err = s.dbMgr.GetByShortID(context.Background(), url /*, userUUID*/)
	}
	if errors.Is(err, ErrDataNotFoundInDB) {
		return "", NewDataNotFoundError(ErrDataNotFoundInDB)
	} else if err != nil {
		return "", fmt.Errorf("failed to retrieve bind by url `%s` from db: %w", url, err)
	}
	return url, nil
}

func (s *DBURLStorage) Set(r *model.URLStorageRecord) error {
	if err := s.dbMgr.Persist(context.Background(), r); err != nil {
		return fmt.Errorf("failed to persist record to db: %w", err)
	}
	return nil
}

func (s *DBURLStorage) BatchSet(binds *[]model.URLStorageRecord) error {
	if err := s.dbMgr.PersistBatch(context.Background(), binds); err != nil {
		return fmt.Errorf("failed to persist batch records to db: %w", err)
	}
	return nil
}

func (s *DBURLStorage) GetByUserUUID(userUUID string) (*[]model.URLStorageRecord, error) {
	urls, err := s.dbMgr.GetByUserUUID(context.Background(), userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user urls from db: %w", err)
	}
	return urls, nil
}
