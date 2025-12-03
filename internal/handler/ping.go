package handler

import (
	"net/http"

	"go.uber.org/zap"
)

// PingProcessor defines the interface for processing service health check requests.
// Implementations handle the business logic of checking service readiness and connectivity.
type PingProcessor interface {
	Process() error
}

// HandlePing creates an HTTP handler for service health checks.
// It handles GET requests to '/ping' endpoint to verify service availability.
//
// The handler:
//   - Performs a health check by pinging the service dependencies
//   - Returns appropriate HTTP status codes:
//   - 200 OK when service is healthy and responsive
//   - 500 Internal Server Error when service is unavailable
//
// Parameters:
//   - p: Processor implementing the health check logic
//   - l: Logger for logging operations
//
// Returns:
//   - HTTP handler function for the ping endpoint
func HandlePing(p PingProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		err := p.Process()
		if err != nil {
			l.Error("failed to ping service", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
