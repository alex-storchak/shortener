package repository

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

type dbManager interface {
	GetByOriginalURL(ctx context.Context, origURL string) (string, error)
	GetByShortID(ctx context.Context, shortID string) (string, error)
	Persist(ctx context.Context, origURL, shortID string) error
}

type DBURLStorage struct {
	logger *zap.Logger
	dbMgr  dbManager
}

func NewDBURLStorage(logger *zap.Logger, dbm dbManager) *DBURLStorage {
	logger = logger.With(
		zap.String("component", "storage"),
	)
	return &DBURLStorage{
		logger: logger,
		dbMgr:  dbm,
	}
}

func (s *DBURLStorage) Get(url, searchByType string) (string, error) {
	s.logger.Debug("Getting url from storage by type",
		zap.String("url", url),
		zap.String("searchByType", searchByType),
	)

	var err error
	if searchByType == OrigURLType {
		url, err = s.dbMgr.GetByOriginalURL(context.Background(), url)
	} else if searchByType == ShortURLType {
		url, err = s.dbMgr.GetByShortID(context.Background(), url)
	}
	if errors.Is(err, ErrDataNotFoundInDB) {
		return "", ErrURLStorageDataNotFound
	} else if err != nil {
		return "", err
	}
	return url, nil
}

func (s *DBURLStorage) Set(origURL, shortURL string) error {
	s.logger.Debug("Setting url to storage",
		zap.String("originalURL", origURL),
		zap.String("shortURL", shortURL),
	)
	if err := s.dbMgr.Persist(context.Background(), origURL, shortURL); err != nil {
		s.logger.Error("Can't persist record to db", zap.Error(err))
		return err
	}
	return nil
}
