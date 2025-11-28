package processor

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type APIUserURLs struct {
	shortener service.URLShortener
	logger    *zap.Logger
	ub        ShortURLBuilder
}

func NewAPIUserURLs(
	shortener service.URLShortener,
	logger *zap.Logger,
	ub ShortURLBuilder,
) *APIUserURLs {
	return &APIUserURLs{
		shortener: shortener,
		logger:    logger,
		ub:        ub,
	}
}

func (s *APIUserURLs) ProcessGet(ctx context.Context) ([]model.UserURLsResponseItem, error) {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user uuid from context: %w", err)
	}
	urls, err := s.shortener.GetUserURLs(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("get user urls from storage: %w", err)
	}

	resp, err := s.buildResponse(urls)
	if err != nil {
		return nil, fmt.Errorf("build response: %w", err)
	}

	return resp, nil
}

func (s *APIUserURLs) buildResponse(urls []*model.URLStorageRecord) ([]model.UserURLsResponseItem, error) {
	resp := make([]model.UserURLsResponseItem, len(urls))
	for i, u := range urls {
		shortURL := s.ub.Build(u.ShortID)
		resp[i] = model.UserURLsResponseItem{
			OrigURL:  u.OrigURL,
			ShortURL: shortURL,
		}
	}
	return resp, nil
}

func (s *APIUserURLs) ProcessDelete(ctx context.Context, shortIDs []string) error {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return fmt.Errorf("get user uuid from context: %w", err)
	}

	go s.processDeletion(context.Background(), userUUID, shortIDs)

	return nil
}

func (s *APIUserURLs) processDeletion(
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
	s.deleteBatch(ctx, batchesCh)
}

func (s *APIUserURLs) fanOut(ctx context.Context, inputCh chan model.URLToDelete) []chan model.URLDeleteBatch {
	numWorkers := runtime.NumCPU()
	channels := make([]chan model.URLDeleteBatch, numWorkers)
	batchSize := 50
	for i := 0; i < numWorkers; i++ {
		outCh := s.batchWorker(ctx, inputCh, batchSize)
		channels[i] = outCh
	}
	return channels
}

func (s *APIUserURLs) batchWorker(
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

func (s *APIUserURLs) fanIn(ctx context.Context, channels ...chan model.URLDeleteBatch) chan model.URLDeleteBatch {
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

func (s *APIUserURLs) deleteBatch(ctx context.Context, batchesCh chan model.URLDeleteBatch) {
	for batch := range batchesCh {
		if err := s.shortener.DeleteBatch(ctx, batch); err != nil {
			s.logger.Error("error deleting batch", zap.Error(err))
			return
		}
		s.logger.Debug("batch deleted successfully")
	}
}
