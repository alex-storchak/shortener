package handler

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/codec"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

// APIShortenBatchProcessor defines the interface for processing batch URL shortening requests.
// Implementations handle the business logic of converting multiple original URLs to short URLs
// in a single operation while maintaining correlation between requests and responses.
type APIShortenBatchProcessor interface {
	Process(ctx context.Context, items model.BatchShortenRequest) (model.BatchShortenResponse, error)
}

// HandleAPIShortenBatch creates an HTTP handler for the batch URL shortening API endpoint.
// It handles POST requests to '/api/shorten/batch' with JSON content containing multiple URLs.
//
// The handler:
//   - Validates that Content-Type is 'application/json'
//   - Processes the batch shortening request
//   - Returns appropriate HTTP status codes:
//   - 400 Bad Request for invalid content type, malformed JSON, or empty input
//   - 201 Created with BatchShortenResponse for successful processing
//   - 500 Internal Server Error for internal processing failures
//
// Parameters:
//   - p: Processor implementing the batch shortening business logic
//   - l: Logger for error logging and debugging
//
// Returns:
//   - HTTP handler function for the batch shorten endpoint
func HandleAPIShortenBatch(p APIShortenBatchProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if err := validateContentType(ct, "application/json"); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var req model.BatchShortenRequest
		if err := codec.EasyJSONDecode(r, &req); err != nil {
			l.Debug("decode json request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		respItems, err := p.Process(r.Context(), req)
		if errors.Is(err, service.ErrEmptyInputURL) || errors.Is(err, service.ErrEmptyInputBatch) {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if err != nil {
			l.Error("failed to shorten batch", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = codec.EasyJSONEncode(w, http.StatusCreated, respItems); err != nil {
			l.Error("encode json response", zap.Error(err))
			return
		}
	}
}
