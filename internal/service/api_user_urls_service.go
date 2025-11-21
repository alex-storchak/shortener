package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/alex-storchak/shortener/internal/helper"
	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type DeleteBatchRequestDecoder interface {
	Decode(io.Reader) ([]string, error)
}

type APIUserURLsService struct {
	baseURL   string
	shortener URLShortener
	dec       DeleteBatchRequestDecoder
	logger    *zap.Logger
}

func NewAPIUserURLsService(
	baseURL string,
	shortener URLShortener,
	dec DeleteBatchRequestDecoder,
	logger *zap.Logger,
) *APIUserURLsService {
	return &APIUserURLsService{
		baseURL:   baseURL,
		shortener: shortener,
		dec:       dec,
		logger:    logger,
	}
}

func (s *APIUserURLsService) ProcessGet(ctx context.Context) ([]model.UserURLsResponseItem, error) {
	userUUID, err := helper.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user uuid from context: %w", err)
	}
	urls, err := s.shortener.GetUserURLs(userUUID)
	if err != nil {
		return nil, fmt.Errorf("get user urls from storage: %w", err)
	}

	resp, err := s.buildResponse(urls)
	if err != nil {
		return nil, fmt.Errorf("build response: %w", err)
	}

	return resp, nil
}

func (s *APIUserURLsService) buildResponse(urls []*model.URLStorageRecord) ([]model.UserURLsResponseItem, error) {
	resp := make([]model.UserURLsResponseItem, len(urls))
	for i, u := range urls {
		shortURL, err := url.JoinPath(s.baseURL, u.ShortID)
		if err != nil {
			return nil, fmt.Errorf("build full short url for new url: %w", err)
		}
		resp[i] = model.UserURLsResponseItem{
			OrigURL:  u.OrigURL,
			ShortURL: shortURL,
		}
	}
	return resp, nil
}

func (s *APIUserURLsService) ProcessDelete(ctx context.Context, r io.Reader) error {
	userUUID, err := helper.GetCtxUserUUID(ctx)
	if err != nil {
		return fmt.Errorf("get user uuid from context: %w", err)
	}
	shortIDs, err := s.dec.Decode(r)
	if err != nil {
		return fmt.Errorf("decode delete batch request json: %w", err)
	}

	go s.processDeletion(context.Background(), userUUID, shortIDs)

	return nil
}

func (s *APIUserURLsService) processDeletion(
	ctx context.Context,
	userUUID string,
	shortIDs []string,
) {
	inputCh := make(chan model.URLToDelete, len(shortIDs))
	go func() {
		defer close(inputCh)
		for _, shortID := range shortIDs {
			select {
			case <-ctx.Done():
				return
			case inputCh <- model.URLToDelete{UserUUID: userUUID, ShortID: shortID}:
			}
		}
	}()

	batchChannels := s.fanOut(ctx, inputCh)
	batchesCh := s.fanIn(ctx, batchChannels...)
	s.deleteBatch(batchesCh)
}

func (s *APIUserURLsService) fanOut(ctx context.Context, inputCh chan model.URLToDelete) []chan model.URLDeleteBatch {
	numWorkers := runtime.NumCPU()
	channels := make([]chan model.URLDeleteBatch, numWorkers)
	batchSize := 50
	for i := 0; i < numWorkers; i++ {
		outCh := s.batchWorker(ctx, inputCh, batchSize)
		channels[i] = outCh
	}
	return channels
}

func (s *APIUserURLsService) batchWorker(
	ctx context.Context,
	inCh chan model.URLToDelete,
	batchSize int,
) chan model.URLDeleteBatch {
	outCh := make(chan model.URLDeleteBatch)

	go func() {
		defer close(outCh)
		batch := make(model.URLDeleteBatch, 0, batchSize)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				flush(ctx, outCh, &batch, batchSize)
				return
			case urlToDelete, ok := <-inCh:
				if !ok {
					flush(ctx, outCh, &batch, batchSize)
					return
				}
				batch = append(batch, urlToDelete)
				if len(batch) >= batchSize {
					flush(ctx, outCh, &batch, batchSize)
				}
			case <-ticker.C:
				flush(ctx, outCh, &batch, batchSize)
			}
		}
	}()

	return outCh
}

func flush(
	ctx context.Context,
	outputCh chan model.URLDeleteBatch,
	batch *model.URLDeleteBatch,
	batchSize int,
) {
	if len(*batch) > 0 {
		select {
		case <-ctx.Done():
		case outputCh <- *batch:
			*batch = make(model.URLDeleteBatch, 0, batchSize)
		}
	}
}

func (s *APIUserURLsService) fanIn(ctx context.Context, channels ...chan model.URLDeleteBatch) chan model.URLDeleteBatch {
	finalCh := make(chan model.URLDeleteBatch)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(inCh chan model.URLDeleteBatch) {
			defer wg.Done()

			for batch := range inCh {
				select {
				case <-ctx.Done():
					return
				case finalCh <- batch:
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}

func (s *APIUserURLsService) deleteBatch(batchesCh chan model.URLDeleteBatch) {
	for batch := range batchesCh {
		if err := s.shortener.DeleteBatch(batch); err != nil {
			s.logger.Error("error deleting batch", zap.Error(err))
			return
		}
		s.logger.Debug("batch deleted successfully")
	}
}
