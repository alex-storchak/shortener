package service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/alex-storchak/shortener/internal/helper"
	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type IAPIUserURLsService interface {
	GetUserURLs(ctx context.Context) (*[]model.UserURLsResponseItem, error)
}

type APIUserURLsService struct {
	baseURL   string
	shortener IShortener
	logger    *zap.Logger
}

func NewAPIUserURLsService(
	baseURL string,
	shortener IShortener,
	logger *zap.Logger,
) *APIUserURLsService {
	return &APIUserURLsService{
		baseURL:   baseURL,
		shortener: shortener,
		logger:    logger,
	}
}

func (s *APIUserURLsService) GetUserURLs(ctx context.Context) (*[]model.UserURLsResponseItem, error) {
	userUUID, err := helper.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user uuid from context: %w", err)
	}
	urls, err := s.shortener.GetUserURLs(userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user urls from storage: %w", err)
	}

	resp, err := s.buildResponse(urls)
	if err != nil {
		return nil, fmt.Errorf("failed to build response: %w", err)
	}

	return &resp, nil
}

func (s *APIUserURLsService) buildResponse(urls *[]model.URLStorageRecord) ([]model.UserURLsResponseItem, error) {
	resp := make([]model.UserURLsResponseItem, len(*urls))
	for i, u := range *urls {
		shortURL, err := url.JoinPath(s.baseURL, u.ShortID)
		if err != nil {
			return nil, fmt.Errorf("failed to build full short url for new url: %w", err)
		}
		resp[i] = model.UserURLsResponseItem{
			OrigURL:  u.OrigURL,
			ShortURL: shortURL,
		}
	}
	return resp, nil
}
