package service

import (
	"errors"

	"go.uber.org/zap"
)

type IMainPageService interface {
	Shorten(body []byte) (shortURL string, err error)
}

type MainPageService struct {
	core   IShortenCore
	logger *zap.Logger
}

func NewMainPageService(core IShortenCore, logger *zap.Logger) *MainPageService {
	return &MainPageService{
		core:   core,
		logger: logger,
	}
}

func (s *MainPageService) Shorten(body []byte) (string, error) {
	shortURL, _, err := s.core.Shorten(string(body))
	if errors.Is(err, ErrEmptyInputURL) {
		return "", ErrEmptyBody
	} else if errors.Is(err, ErrURLAlreadyExists) {
		return shortURL, err
	} else if err != nil {
		return "", err
	}
	return shortURL, nil
}

var (
	ErrEmptyBody = errors.New("request body is empty")
)
