package repository

import "errors"

type URLStorage interface {
	Has(url string) bool
	Get(url string) (string, error)
	Set(url string, shortURL string) error
}

type MapURLStorage struct {
	storage map[string]string
}

func NewMapURLStorage() *MapURLStorage {
	return &MapURLStorage{
		storage: make(map[string]string),
	}
}

func (s *MapURLStorage) Has(url string) bool {
	_, ok := s.storage[url]
	return ok
}

func (s *MapURLStorage) Get(url string) (string, error) {
	value, ok := s.storage[url]
	if !ok {
		return "", ErrURLStorageDataNotFound
	}
	return value, nil
}

func (s *MapURLStorage) Set(keyURL, valueURL string) error {
	s.storage[keyURL] = valueURL
	return nil
}

var (
	ErrURLStorageDataNotFound = errors.New("no data in the storage for the requested url")
)
