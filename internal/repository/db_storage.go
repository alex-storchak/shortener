package repository

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

type dbManager interface {
	GetByOriginalURL(ctx context.Context, origURL string) (string, error)
	GetByShortID(ctx context.Context, shortID string) (string, error)
	Persist(ctx context.Context, origURL, shortID string) error
	PersistBatch(ctx context.Context, binds *[]URLBind) error
	Ping(ctx context.Context) error
	Close() error
}

type DBURLStorage struct {
	logger *zap.Logger
	dbMgr  dbManager
}

func NewDBURLStorage(logger *zap.Logger, dbm dbManager) *DBURLStorage {
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

func (s *DBURLStorage) Get(url, searchByType string) (string, error) {
	var err error
	if searchByType == OrigURLType {
		url, err = s.dbMgr.GetByOriginalURL(context.Background(), url)
	} else if searchByType == ShortURLType {
		url, err = s.dbMgr.GetByShortID(context.Background(), url)
	}
	if errors.Is(err, ErrDataNotFoundInDB) {
		return "", NewDataNotFoundError(ErrDataNotFoundInDB)
	} else if err != nil {
		return "", fmt.Errorf("failed to retrieve bind by url `%s` from db: %w", url, err)
	}
	return url, nil
}

func (s *DBURLStorage) Set(origURL, shortURL string) error {
	if err := s.dbMgr.Persist(context.Background(), origURL, shortURL); err != nil {
		return fmt.Errorf("failed to persist record to db: %w", err)
	}
	return nil
}

func (s *DBURLStorage) BatchSet(binds *[]URLBind) error {
	if err := s.dbMgr.PersistBatch(context.Background(), binds); err != nil {
		return fmt.Errorf("failed to persist batch records to db: %w", err)
	}
	return nil
}
