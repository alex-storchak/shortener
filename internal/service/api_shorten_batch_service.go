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

type ShortenBatchRequestDecoder interface {
	Decode(io.Reader) ([]model.BatchShortenRequestItem, error)
}

type ShortenBatchService struct {
	baseURL   string
	shortener URLShortener
	dec       ShortenBatchRequestDecoder
	logger    *zap.Logger
}

func NewShortenBatchService(
	baseURL string,
	shortener URLShortener,
	dec ShortenBatchRequestDecoder,
	l *zap.Logger,
) *ShortenBatchService {
	return &ShortenBatchService{
		baseURL:   baseURL,
		shortener: shortener,
		dec:       dec,
		logger:    l,
	}
}

func (s *ShortenBatchService) Process(ctx context.Context, r io.Reader) ([]model.BatchShortenResponseItem, error) {
	userUUID, err := helper.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user uuid from context: %w", err)
	}
	reqItems, err := s.dec.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decode shorten batch request json: %w", err)
	}

	origURLs := s.buildURLList(reqItems)
	shortIDs, err := s.shortener.ShortenBatch(userUUID, origURLs)
	if err != nil {
		return nil, fmt.Errorf("shorten batch: %w", err)
	}

	resp, err := s.buildResponse(reqItems, shortIDs)
	if err != nil {
		return nil, fmt.Errorf("build response: %w", err)
	}

	return resp, nil
}

func (s *ShortenBatchService) buildURLList(reqItems []model.BatchShortenRequestItem) []string {
	origURLs := make([]string, len(reqItems))
	for i, item := range reqItems {
		origURLs[i] = item.OriginalURL
	}
	return origURLs
}

func (s *ShortenBatchService) buildResponse(
	reqItems []model.BatchShortenRequestItem,
	shortIDs []string,
) ([]model.BatchShortenResponseItem, error) {
	resp := make([]model.BatchShortenResponseItem, len(reqItems))
	for i, item := range reqItems {
		shortURL, err := url.JoinPath(s.baseURL, shortIDs[i])
		if err != nil {
			return nil, fmt.Errorf("build full short url for new url: %w", err)
		}
		resp[i] = model.BatchShortenResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		}
	}
	return resp, nil
}
