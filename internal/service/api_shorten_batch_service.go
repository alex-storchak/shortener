package service

import (
	"errors"
	"fmt"
	"io"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type IAPIShortenBatchService interface {
	ShortenBatch(r io.Reader) ([]model.BatchShortenResponseItem, error)
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
	logger = logger.With(zap.String("package", "api_shorten_batch_service"))
	return &APIShortenBatchService{
		baseURL:   baseURL,
		shortener: shortener,
		batchDec:  decoder,
		logger:    logger,
	}
}

func (s *APIShortenBatchService) ShortenBatch(r io.Reader) ([]model.BatchShortenResponseItem, error) {
	reqItems, err := s.batchDec.DecodeBatch(r)
	if err != nil {
		s.logger.Debug("cannot decode api/shorten/batch request JSON")
		return nil, ErrJSONDecode
	}
	if len(*reqItems) == 0 {
		return nil, ErrEmptyBatch
	}

	origURLs, err := s.buildURLList(reqItems)
	if err != nil {
		return nil, err
	}

	shortIDs, err := s.shortener.ShortenBatch(origURLs)
	if errors.Is(err, ErrEmptyInputURL) {
		return nil, ErrEmptyURL
	} else if err != nil {
		return nil, err
	}

	resp := s.buildResponse(reqItems, shortIDs)

	return resp, nil
}

func (s *APIShortenBatchService) buildURLList(reqItems *[]model.BatchShortenRequestItem) (*[]string, error) {
	origURLs := make([]string, len(*reqItems))
	for i, item := range *reqItems {
		if item.OriginalURL == "" {
			return nil, ErrEmptyURL
		}
		origURLs[i] = item.OriginalURL
	}
	return &origURLs, nil
}

func (s *APIShortenBatchService) buildResponse(
	reqItems *[]model.BatchShortenRequestItem,
	shortIDs *[]string,
) []model.BatchShortenResponseItem {
	resp := make([]model.BatchShortenResponseItem, len(*reqItems))
	for i, item := range *reqItems {
		resp[i] = model.BatchShortenResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.baseURL, (*shortIDs)[i]),
		}
	}
	return resp
}

var (
	ErrEmptyBatch = errors.New("empty batch provided")
)
