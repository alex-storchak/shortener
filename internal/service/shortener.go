package service

import (
	"errors"

	"github.com/alex-storchak/shortener/internal/repository"
	"go.uber.org/zap"
)

type IShortener interface {
	Shorten(url string) (shortID string, err error)
	Extract(shortID string) (OrigURL string, err error)
	ShortenBatch(urls *[]string) (*[]string, error)
}

type Shortener struct {
	urlStorage repository.URLStorage
	generator  IDGenerator
	logger     *zap.Logger
}

func NewShortener(
	idGenerator IDGenerator,
	urlStorage repository.URLStorage,
	logger *zap.Logger,
) *Shortener {
	logger = logger.With(
		zap.String("package", "shortener"),
	)
	return &Shortener{
		urlStorage: urlStorage,
		generator:  idGenerator,
		logger:     logger,
	}
}

func (s *Shortener) Shorten(url string) (string, error) {
	shortID, err := s.urlStorage.Get(url, repository.OrigURLType)
	if err == nil {
		s.logger.Debug("url already exists in the storage", zap.String("url", url))
		return shortID, ErrURLAlreadyExists
	} else if !errors.Is(err, repository.ErrURLStorageDataNotFound) {
		s.logger.Error("error retrieving url", zap.Error(err))
		return "", err
	}

	shortID, err = s.generator.Generate()
	s.logger.Debug("generated short id", zap.String("shortID", shortID))
	if err != nil {
		s.logger.Error("failed to generate short id", zap.Error(err))
		return "", ErrShortenerGenerationShortIDFailed
	}
	if err := s.urlStorage.Set(url, shortID); err != nil {
		s.logger.Error("failed to set url binding in the urlStorage", zap.Error(err))
		return "", ErrShortenerSetBindingURLStorageFailed
	}
	return shortID, nil
}

func (s *Shortener) Extract(shortID string) (string, error) {
	origURL, err := s.urlStorage.Get(shortID, repository.ShortURLType)
	if err != nil {
		return "", err
	}
	return origURL, nil
}

func (s *Shortener) ShortenBatch(urls *[]string) (*[]string, error) {
	res, toPersist, err := s.segregateBatch(urls)
	if err != nil {
		return nil, err
	}

	if len(*toPersist) > 0 {
		if err := s.urlStorage.BatchSet(toPersist); err != nil {
			s.logger.Error("failed to set url bindings batch in the urlStorage", zap.Error(err))
			return nil, ErrShortenerSetBindingURLStorageFailed
		}
	}

	return res, nil
}

func (s *Shortener) segregateBatch(urls *[]string) (*[]string, *[]repository.URLBind, error) {
	res := make([]string, len(*urls))
	toPersist := make([]repository.URLBind, 0)

	for i, u := range *urls {
		if u == "" {
			return nil, nil, ErrEmptyInputURL
		}

		shortID, err := s.urlStorage.Get(u, repository.OrigURLType)
		if err == nil {
			res[i] = shortID
			continue
		} else if !errors.Is(err, repository.ErrURLStorageDataNotFound) {
			s.logger.Error("error retrieving url", zap.Error(err))
			return nil, nil, err
		}

		urlBindItem, err := s.prepareURLBindToPersistItem(u)
		if err != nil {
			return nil, nil, err
		}
		toPersist = append(toPersist, urlBindItem)
		res[i] = urlBindItem.ShortID
	}
	return &res, &toPersist, nil
}

func (s *Shortener) prepareURLBindToPersistItem(origURL string) (repository.URLBind, error) {
	shortID, err := s.generator.Generate()
	s.logger.Debug("generated short id", zap.String("shortID", shortID))
	if err != nil {
		s.logger.Error("failed to generate short id", zap.Error(err))
		return repository.URLBind{}, ErrShortenerGenerationShortIDFailed
	}
	return repository.URLBind{OrigURL: origURL, ShortID: shortID}, nil
}

var (
	ErrShortenerGenerationShortIDFailed    = errors.New("failed to generate short id")
	ErrShortenerSetBindingURLStorageFailed = errors.New("failed to set url binding in the urlStorage")
	ErrURLAlreadyExists                    = errors.New("url already exists")
)
