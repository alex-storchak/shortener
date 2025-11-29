package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseData holds captured response metrics for logging.
type responseData struct {
	httpStatus int
	size       int
}

// loggingResponseWriter wraps http.ResponseWriter to capture response status and size.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
	logger       *zap.Logger
}

// Write captures the number of bytes written and delegates to the underlying ResponseWriter.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader captures the HTTP status code and delegates to the underlying ResponseWriter.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.httpStatus = statusCode
}

// newLoggingResponseWriter creates a new logging response writer.
func newLoggingResponseWriter(w http.ResponseWriter, rd *responseData, logger *zap.Logger) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		responseData:   rd,
		logger:         logger,
	}
}

// logSummary logs request details including method, URI, status, size, and duration.
func logSummary(logger *zap.Logger, r *http.Request, rd *responseData, start time.Time) {
	logger.Info("got incoming HTTP request",
		zap.String("method", r.Method),
		zap.String("URI", r.RequestURI),
		zap.Int("status", rd.httpStatus),
		zap.Int("size", rd.size),
		zap.String("duration", time.Since(start).String()),
	)
}

// NewRequestLogger creates middleware that logs all HTTP requests with detailed metrics.
// The middleware captures:
//   - HTTP method and request URI
//   - Response status code
//   - Response body size in bytes
//   - Request processing duration
//
// Parameters:
//   - logger: structured logger for logging operations
//
// Returns:
//   - func(http.Handler) http.Handler: request logging middleware function
func NewRequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rd := &responseData{}
			lw := newLoggingResponseWriter(w, rd, logger)
			next.ServeHTTP(lw, r)

			logSummary(logger, r, rd, start)
		})
	}
}
