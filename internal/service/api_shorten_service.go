package service

import (
	"errors"
	"io"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type IAPIShortenService interface {
	Shorten(r io.Reader) (*model.ShortenResponse, error)
}

type APIShortenService struct {
	core    IShortenCore
	decoder IJSONRequestDecoder
	logger  *zap.Logger
}

func NewAPIShortenService(core IShortenCore, decoder IJSONRequestDecoder, logger *zap.Logger) *APIShortenService {
	return &APIShortenService{
		core:    core,
		decoder: decoder,
		logger:  logger,
	}
}

func (s *APIShortenService) Shorten(r io.Reader) (*model.ShortenResponse, error) {
	req, err := s.decoder.Decode(r)
	if err != nil {
		return nil, ErrJSONDecode
	}

	shortURL, _, err := s.core.Shorten(req.OrigURL)
	if errors.Is(err, ErrEmptyInputURL) {
		return nil, ErrEmptyURL
	} else if errors.Is(err, ErrURLAlreadyExists) {
		return &model.ShortenResponse{ShortURL: shortURL}, ErrURLAlreadyExists
	} else if err != nil {
		return nil, err
	}
	return &model.ShortenResponse{ShortURL: shortURL}, nil
}

var (
	ErrEmptyURL   = errors.New("empty url provided")
	ErrJSONDecode = errors.New("failed to decode json")
)
