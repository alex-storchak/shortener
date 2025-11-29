package processor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

type APIShortenBatch struct {
	shortener service.URLShortener
	logger    *zap.Logger
	ub        ShortURLBuilder
}

func NewAPIShortenBatch(s service.URLShortener, l *zap.Logger, ub ShortURLBuilder) *APIShortenBatch {
	return &APIShortenBatch{
		shortener: s,
		logger:    l,
		ub:        ub,
	}
}

func (s *APIShortenBatch) Process(ctx context.Context, items model.BatchShortenRequest) (model.BatchShortenResponse, error) {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user uuid from context: %w", err)
	}

	origURLs := s.buildURLList(items)
	shortIDs, err := s.shortener.ShortenBatch(ctx, userUUID, origURLs)
	if err != nil {
		return nil, fmt.Errorf("shorten batch: %w", err)
	}

	resp, err := s.buildResponse(items, shortIDs)
	if err != nil {
		return nil, fmt.Errorf("build response: %w", err)
	}

	return resp, nil
}

func (s *APIShortenBatch) buildURLList(reqItems []model.BatchShortenRequestItem) []string {
	origURLs := make([]string, len(reqItems))
	for i, item := range reqItems {
		origURLs[i] = item.OriginalURL
	}
	return origURLs
}

func (s *APIShortenBatch) buildResponse(
	reqItems []model.BatchShortenRequestItem,
	shortIDs []string,
) ([]model.BatchShortenResponseItem, error) {
	resp := make([]model.BatchShortenResponseItem, len(reqItems))
	for i, item := range reqItems {
		shortURL := s.ub.Build(shortIDs[i])
		resp[i] = model.BatchShortenResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		}
	}
	return resp, nil
}
