package repository

import "errors"

type UrlStorage interface {
	Has(url string) bool
	Get(url string) (string, error)
	Set(url string, shortUrl string) error
}

type MapUrlStorage struct {
	storage map[string]string
}

func NewMapUrlStorage() MapUrlStorage {
	return MapUrlStorage{
		storage: make(map[string]string),
	}
}

func (s MapUrlStorage) Has(url string) bool {
	_, ok := s.storage[url]
	return ok
}

func (s MapUrlStorage) Get(url string) (string, error) {
	value, ok := s.storage[url]
	if !ok {
		return "", ErrShortUrlNotFound
	}
	return value, nil
}

func (s MapUrlStorage) Set(url string, shortUrl string) error {
	s.storage[url] = shortUrl
	return nil
}

var (
	ErrShortUrlNotFound = errors.New("no data in the storage for the requested url")
)
