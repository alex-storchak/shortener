package service

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/alex-storchak/shortener/internal/helper"
	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type IAPIShortenBatchService interface {
	ShortenBatch(ctx context.Context, r io.Reader) ([]model.BatchShortenResponseItem, error)
}

type APIShortenBatchService struct {
	baseURL   string
	shortener IShortener
	batchDec  IJSONBatchRequestDecoder
	logger    *zap.Logger
}

func NewAPIShortenBatchService(
	baseURL string,
	shortener IShortener,
	decoder IJSONBatchRequestDecoder,
	logger *zap.Logger,
) *APIShortenBatchService {
	return &APIShortenBatchService{
		baseURL:   baseURL,
		shortener: shortener,
		batchDec:  decoder,
		logger:    logger,
	}
}

func (s *APIShortenBatchService) ShortenBatch(ctx context.Context, r io.Reader) ([]model.BatchShortenResponseItem, error) {
	userUUID, err := helper.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user uuid from context: %w", err)
	}
	reqItems, err := s.batchDec.DecodeBatch(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode batch request json: %w", err)
	}

	origURLs := s.buildURLList(reqItems)
	shortIDs, err := s.shortener.ShortenBatch(userUUID, origURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to shorten batch: %w", err)
	}

	resp, err := s.buildResponse(reqItems, shortIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to build response: %w", err)
	}

	return resp, nil
}

func (s *APIShortenBatchService) buildURLList(reqItems []model.BatchShortenRequestItem) []string {
	origURLs := make([]string, len(reqItems))
	for i, item := range reqItems {
		origURLs[i] = item.OriginalURL
	}
	return origURLs
}

func (s *APIShortenBatchService) buildResponse(
	reqItems []model.BatchShortenRequestItem,
	shortIDs []string,
) ([]model.BatchShortenResponseItem, error) {
	resp := make([]model.BatchShortenResponseItem, len(reqItems))
	for i, item := range reqItems {
		shortURL, err := url.JoinPath(s.baseURL, shortIDs[i])
		if err != nil {
			return nil, fmt.Errorf("failed to build full short url for new url: %w", err)
		}
		resp[i] = model.BatchShortenResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		}
	}
	return resp, nil
}
