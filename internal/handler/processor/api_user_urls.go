package processor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

// APIUserURLs provides user URL management functionality.
// It handles business logic for user-specific URL operations including retrieval and deletion.
type APIUserURLs struct {
	shortener service.URLShortener
	logger    *zap.Logger
	ub        ShortURLBuilder
}

// NewAPIUserURLs creates a new APIUserURLs processor instance.
//
// Parameters:
//   - shortener: URL shortener service for user URL operations
//   - logger: Structured logger for logging operations
//   - ub: URL builder for constructing complete short URLs
//
// Returns: configured APIUserURLs processor
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

// ProcessGet retrieves all URLs shortened by the authenticated user.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//
// Returns:
//   - model.UserURLsGetResponse: collection of user's shortened URLs
//   - error: nil on success, or service error if operation fails
func (s *APIUserURLs) ProcessGet(ctx context.Context) (model.UserURLsGetResponse, error) {
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

// buildResponse creates a response with user's shortened URLs.
//
// Parameters:
//   - urls: List of model.URLStorageRecord containing user's shortened URLs
//
// Returns:
//   - model.UserURLsGetResponse: Collection of user's shortened URLs
//   - error: nil on success, or error if response construction fails
func (s *APIUserURLs) buildResponse(urls []*model.URLStorageRecord) (model.UserURLsGetResponse, error) {
	resp := make(model.UserURLsGetResponse, len(urls))
	for i, u := range urls {
		shortURL := s.ub.Build(u.ShortID)
		resp[i] = model.UserURLsGetResponseItem{
			OrigURL:  u.OrigURL,
			ShortURL: shortURL,
		}
	}
	return resp, nil
}

// ProcessDelete initiates asynchronous batch deletion of user's URLs.
// The deletion is processed in the background using a fan-out/fan-in pattern for efficiency.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - shortIDs: list of short URL identifiers to delete
//
// Returns:
//   - error: nil if deletion request was accepted, or error if authentication fails
//
// Note: Actual deletion happens asynchronously, method returns immediately after validation.
func (s *APIUserURLs) ProcessDelete(ctx context.Context, shortIDs model.UserURLsDelRequest) error {
	userUUID, err := auth.GetCtxUserUUID(ctx)
	if err != nil {
		return fmt.Errorf("get user uuid from context: %w", err)
	}

	go s.processDeletion(context.Background(), userUUID, shortIDs)

	return nil
}

// processDeletion handles the deletion of user's URLs in the background.
// It uses a fan-out/fan-in pattern to process deletions in parallel.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - userUUID: UUID of the user whose URLs are being deleted
//   - shortIDs: list of short URL identifiers to delete
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

// fanOut distributes the input channel into multiple worker channels.
// It creates a fixed number of worker channels and assigns input items to them.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - inputCh: channel of model.URLToDelete to be distributed
//
// Returns:
//   - []chan model.URLDeleteBatch: list of worker channels
func (s *APIUserURLs) fanOut(ctx context.Context, inputCh chan model.URLToDelete) []chan model.URLDeleteBatch {
	const numWorkers = 4
	channels := make([]chan model.URLDeleteBatch, numWorkers)
	batchSize := 50
	for i := 0; i < numWorkers; i++ {
		outCh := s.batchWorker(ctx, inputCh, batchSize)
		channels[i] = outCh
	}
	return channels
}

// batchWorker processes a batch of URLs to delete.
// It accumulates URLs in a batch and flushes them to the output channel
// at regular intervals or when the batch size is reached.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - inCh: channel of model.URLToDelete to be processed
//   - batchSize: maximum number of URLs to accumulate in a batch
//
// Returns:
//   - chan model.URLDeleteBatch: channel of batches to be deleted
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

// flush sends the accumulated batch to the output channel.
// If the batch is empty, it does nothing.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - outputCh: channel to send the batch
//   - batch: pointer to the batch to be sent
//   - batchSize: maximum number of URLs in a batch
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

// fanIn merges multiple input channels into a single output channel.
// It waits for all input channels to close before closing the output channel.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - channels: list of input channels to merge
//
// Returns:
//   - chan model.URLDeleteBatch: merged output channel
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

// deleteBatch deletes a batch of URLs from storage.
// It logs any errors that occur during deletion.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - batchesCh: channel of batches to delete
func (s *APIUserURLs) deleteBatch(ctx context.Context, batchesCh chan model.URLDeleteBatch) {
	for batch := range batchesCh {
		if err := s.shortener.DeleteBatch(ctx, batch); err != nil {
			s.logger.Error("error deleting batch", zap.Error(err))
			return
		}
		s.logger.Debug("batch deleted successfully")
	}
}
