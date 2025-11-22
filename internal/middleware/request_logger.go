package middleware

import (
	"net/http"
	"time"

	"github.com/alex-storchak/shortener/internal/helper/datetime"
	"go.uber.org/zap"
)

type responseData struct {
	httpStatus int
	size       int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
	logger       *zap.Logger
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.httpStatus = statusCode
}

func newLoggingResponseWriter(w http.ResponseWriter, rd *responseData, logger *zap.Logger) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		responseData:   rd,
		logger:         logger,
	}
}

func logSummary(logger *zap.Logger, r *http.Request, rd *responseData, start time.Time) {
	duration := time.Since(start)
	fmtDuration := datetime.FormatDuration(duration)

	logger.Info("got incoming HTTP request",
		zap.String("method", r.Method),
		zap.String("URI", r.RequestURI),
		zap.Int("status", rd.httpStatus),
		zap.Int("size", rd.size),
		zap.String("duration", fmtDuration),
	)
}

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
