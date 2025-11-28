package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
	repo "github.com/alex-storchak/shortener/internal/repository"
	"go.uber.org/zap"
)

type IDGenerator interface {
	Generate() (string, error)
}

type Pinger interface {
	IsReady(ctx context.Context) error
}

type URLShortener interface {
	Shorten(ctx context.Context, userUUID string, url string) (shortID string, err error)
	Extract(ctx context.Context, shortID string) (OrigURL string, err error)
	ShortenBatch(ctx context.Context, userUUID string, urls []string) ([]string, error)
	GetUserURLs(ctx context.Context, userUUID string) ([]*model.URLStorageRecord, error)
	DeleteBatch(ctx context.Context, urls model.URLDeleteBatch) error
}

type PingableURLShortener interface {
	URLShortener
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

func (s *Shortener) Shorten(ctx context.Context, userUUID string, url string) (string, error) {
	if len(url) == 0 {
		return "", ErrEmptyInputURL
	}

	var (
		r   *model.URLStorageRecord
		err error
	)
	r, err = s.urlStorage.Get(ctx, url, repo.OrigURLType)
	if err == nil {
		return r.ShortID, ErrURLAlreadyExists
	}
	var nfErr *repo.DataNotFoundError
	if !errors.As(err, &nfErr) {
		return "", fmt.Errorf("retrieve url from storage: %w", err)
	}

	shortID, err := s.generator.Generate()
	if err != nil {
		return "", fmt.Errorf("generate short id: %w", err)
	}
	if err := s.urlStorage.Set(ctx, url, shortID, userUUID); err != nil {
		return "", fmt.Errorf("set url binding in storage: %w", err)
	}
	return shortID, nil
}

func (s *Shortener) Extract(ctx context.Context, shortID string) (string, error) {
	r, err := s.urlStorage.Get(ctx, shortID, repo.ShortURLType)
	if err != nil {
		return "", fmt.Errorf("retrieve short url from storage: %w", err)
	}
	return r.OrigURL, nil
}

func (s *Shortener) ShortenBatch(ctx context.Context, userUUID string, urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, ErrEmptyInputBatch
	}

	res, toPersist, err := s.segregateBatch(ctx, userUUID, urls)
	if err != nil {
		return nil, fmt.Errorf("segregate batch: %w", err)
	}
	if len(toPersist) > 0 {
		if err := s.urlStorage.BatchSet(ctx, toPersist); err != nil {
			return nil, fmt.Errorf("set url bindings batch in storage: %w", err)
		}
	}
	return res, nil
}

func (s *Shortener) segregateBatch(ctx context.Context, userUUID string, urls []string) ([]string, []model.URLStorageRecord, error) {
	res := make([]string, len(urls))
	toPersist := make([]model.URLStorageRecord, 0)

	for i, u := range urls {
		if u == "" {
			return nil, nil, ErrEmptyInputURL
		}

		r, err := s.urlStorage.Get(ctx, u, repo.OrigURLType)
		if err == nil {
			res[i] = r.ShortID
			continue
		}
		var nfErr *repo.DataNotFoundError
		if !errors.As(err, &nfErr) {
			return nil, nil, fmt.Errorf("retrieve url from storage: %w", err)
		}

		urlBindItem, err := s.prepareURLBindToPersistItem(userUUID, u)
		if err != nil {
			return nil, nil, fmt.Errorf("prepare url bind to persist item: %w", err)
		}
		toPersist = append(toPersist, urlBindItem)
		res[i] = urlBindItem.ShortID
	}
	return res, toPersist, nil
}

func (s *Shortener) prepareURLBindToPersistItem(userUUID string, origURL string) (model.URLStorageRecord, error) {
	shortID, err := s.generator.Generate()
	if err != nil {
		return model.URLStorageRecord{}, fmt.Errorf("batch. generate short id: %w", err)
	}
	return model.URLStorageRecord{OrigURL: origURL, ShortID: shortID, UserUUID: userUUID}, nil
}

func (s *Shortener) IsReady(ctx context.Context) error {
	return s.urlStorage.Ping(ctx)
}

func (s *Shortener) GetUserURLs(ctx context.Context, userUUID string) ([]*model.URLStorageRecord, error) {
	return s.urlStorage.GetByUserUUID(ctx, userUUID)
}

func (s *Shortener) DeleteBatch(ctx context.Context, urls model.URLDeleteBatch) error {
	return s.urlStorage.DeleteBatch(ctx, urls)
}

var (
	ErrURLAlreadyExists = errors.New("url already exists")
	ErrEmptyInputURL    = errors.New("empty url in the input")
	ErrEmptyInputBatch  = errors.New("empty batch provided")
)
