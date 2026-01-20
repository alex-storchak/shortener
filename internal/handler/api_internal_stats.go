package handler

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/codec"
	"github.com/alex-storchak/shortener/internal/model"
)

// APIInternalProcessor defines the interface for processing internal API statistics requests.
// Implementations are responsible for collecting and returning service statistics.
type APIInternalProcessor interface {
	Process(ctx context.Context) (model.StatsResponse, error)
}

// HandleStats creates an HTTP handler function for serving statistics endpoints.
// The handler processes requests to retrieve service statistics including total shortened URL and user counts.
//
// Parameters:
//   - p: Processor implementing the APIInternalProcessor interface for statistics retrieval
//   - l: Structured logger for logging errors and debug information
//
// Returns:
// - 500 Internal Server Error if statistics cannot be retrieved
// - 200 OK when statistics successfully retrieved
func HandleStats(p APIInternalProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resBody, err := p.Process(r.Context())
		if err != nil {
			l.Error("Failed to get stats", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err = codec.EasyJSONEncode(w, http.StatusOK, resBody); err != nil {
			l.Error("created. encode json response", zap.Error(err))
			return
		}
	}
}
