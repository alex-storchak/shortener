package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
	repo "github.com/alex-storchak/shortener/internal/repository"
	"go.uber.org/zap"
)

type IShortener interface {
	Shorten(userUUID string, url string) (shortID string, err error)
	Extract(shortID string) (OrigURL string, err error)
	ShortenBatch(userUUID string, urls []string) ([]string, error)
	GetUserURLs(userUUID string) ([]*model.URLStorageRecord, error)
	DeleteBatch(urls model.URLDeleteBatch) error
}

type IShortenerService interface {
	IShortener
	Pinger
}

type Shortener struct {
	urlStorage repo.URLStorage
	generator  IDGenerator
	logger     *zap.Logger
}

func NewShortener(
	idGenerator IDGenerator,
	urlStorage repo.URLStorage,
	logger *zap.Logger,
) *Shortener {
	return &Shortener{
		urlStorage: urlStorage,
		generator:  idGenerator,
		logger:     logger,
	}
}

func (s *Shortener) Shorten(userUUID string, url string) (string, error) {
	if len(url) == 0 {
		return "", ErrEmptyInputURL
	}

	var (
		r   *model.URLStorageRecord
		err error
	)
	r, err = s.urlStorage.Get(url, repo.OrigURLType)
	if err == nil {
		return r.ShortID, ErrURLAlreadyExists
	}
	var nfErr *repo.DataNotFoundError
	if !errors.As(err, &nfErr) {
		return "", fmt.Errorf("failed to retrieve url from storage: %w", err)
	}

	shortID, err := s.generator.Generate()
	if err != nil {
		return "", fmt.Errorf("failed to generate short id: %w", err)
	}
	r = &model.URLStorageRecord{OrigURL: url, ShortID: shortID, UserUUID: userUUID}
	if err := s.urlStorage.Set(r); err != nil {
		return "", fmt.Errorf("failed to set url binding in storage: %w", err)
	}
	return shortID, nil
}

func (s *Shortener) Extract(shortID string) (string, error) {
	r, err := s.urlStorage.Get(shortID, repo.ShortURLType)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve short url from storage: %w", err)
	}
	return r.OrigURL, nil
}

func (s *Shortener) ShortenBatch(userUUID string, urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, ErrEmptyInputBatch
	}

	res, toPersist, err := s.segregateBatch(userUUID, urls)
	if err != nil {
		return nil, fmt.Errorf("failed to segregate batch: %w", err)
	}
	if len(toPersist) > 0 {
		if err := s.urlStorage.BatchSet(toPersist); err != nil {
			return nil, fmt.Errorf("failed to set url bindings batch in storage: %w", err)
		}
	}
	return res, nil
}

func (s *Shortener) segregateBatch(userUUID string, urls []string) ([]string, []*model.URLStorageRecord, error) {
	res := make([]string, len(urls))
	toPersist := make([]*model.URLStorageRecord, 0)

	for i, u := range urls {
		if u == "" {
			return nil, nil, ErrEmptyInputURL
		}

		r, err := s.urlStorage.Get(u, repo.OrigURLType)
		if err == nil {
			res[i] = r.ShortID
			continue
		}
		var nfErr *repo.DataNotFoundError
		if !errors.As(err, &nfErr) {
			return nil, nil, fmt.Errorf("failed to retrieve url from storage: %w", err)
		}

		urlBindItem, err := s.prepareURLBindToPersistItem(userUUID, u)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to prepare url bind to persist item: %w", err)
		}
		toPersist = append(toPersist, &urlBindItem)
		res[i] = urlBindItem.ShortID
	}
	return res, toPersist, nil
}

func (s *Shortener) prepareURLBindToPersistItem(userUUID string, origURL string) (model.URLStorageRecord, error) {
	shortID, err := s.generator.Generate()
	if err != nil {
		return model.URLStorageRecord{}, fmt.Errorf("batch. failed to generate short id: %w", err)
	}
	return model.URLStorageRecord{OrigURL: origURL, ShortID: shortID, UserUUID: userUUID}, nil
}

func (s *Shortener) IsReady(ctx context.Context) error {
	return s.urlStorage.Ping(ctx)
}

func (s *Shortener) GetUserURLs(userUUID string) ([]*model.URLStorageRecord, error) {
	return s.urlStorage.GetByUserUUID(userUUID)
}

func (s *Shortener) DeleteBatch(urls model.URLDeleteBatch) error {
	return s.urlStorage.DeleteBatch(urls)
}

var (
	ErrURLAlreadyExists = errors.New("url already exists")
	ErrEmptyInputURL    = errors.New("empty url in the input")
	ErrEmptyInputBatch  = errors.New("empty batch provided")
)
