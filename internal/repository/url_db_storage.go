package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type URLDBManager interface {
	GetByOriginalURL(ctx context.Context, origURL string) (*model.URLStorageRecord, error)
	GetByShortID(ctx context.Context, shortID string) (*model.URLStorageRecord, error)
	Persist(ctx context.Context, r *model.URLStorageRecord) error
	PersistBatch(ctx context.Context, binds []*model.URLStorageRecord) error
	Ping(ctx context.Context) error
	Close() error
	GetByUserUUID(ctx context.Context, userUUID string) ([]*model.URLStorageRecord, error)
	DeleteBatch(ctx context.Context, urls model.URLDeleteBatch) error
}

type DBURLStorage struct {
	logger *zap.Logger
	dbMgr  URLDBManager
}

func NewDBURLStorage(logger *zap.Logger, dbm URLDBManager) *DBURLStorage {
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

func (s *DBURLStorage) Get(url, searchByType string) (*model.URLStorageRecord, error) {
	var (
		r   *model.URLStorageRecord
		err error
	)
	if searchByType == OrigURLType {
		r, err = s.dbMgr.GetByOriginalURL(context.Background(), url)
	} else if searchByType == ShortURLType {
		r, err = s.dbMgr.GetByShortID(context.Background(), url)
		if err == nil && r.IsDeleted {
			return nil, ErrDataDeleted
		}
	}
	if errors.Is(err, ErrDataNotFoundInDB) {
		return nil, NewDataNotFoundError(ErrDataNotFoundInDB)
	} else if err != nil {
		return nil, fmt.Errorf("failed to retrieve bind by url `%s` from db: %w", url, err)
	}
	return r, nil
}

func (s *DBURLStorage) Set(r *model.URLStorageRecord) error {
	if err := s.dbMgr.Persist(context.Background(), r); err != nil {
		return fmt.Errorf("failed to persist record to db: %w", err)
	}
	return nil
}

func (s *DBURLStorage) BatchSet(binds []*model.URLStorageRecord) error {
	if err := s.dbMgr.PersistBatch(context.Background(), binds); err != nil {
		return fmt.Errorf("failed to persist batch records to db: %w", err)
	}
	return nil
}

func (s *DBURLStorage) GetByUserUUID(userUUID string) ([]*model.URLStorageRecord, error) {
	return s.dbMgr.GetByUserUUID(context.Background(), userUUID)
}

func (s *DBURLStorage) DeleteBatch(urls model.URLDeleteBatch) error {
	return s.dbMgr.DeleteBatch(context.Background(), urls)
}
