package middleware

import (
	"net/http"
	"time"

	"github.com/alex-storchak/shortener/internal/service"
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
	r.logger.Debug("actual and added response size",
		zap.Int("actual", r.responseData.size),
		zap.Int("size", size),
	)
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.logger.Debug("added response status", zap.Int("status", statusCode))
	r.responseData.httpStatus = statusCode
}

func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				httpStatus: 0,
				size:       0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
				logger:         logger,
			}
			next.ServeHTTP(&lw, r)

			duration := time.Since(start)
			fmtDuration := service.FormatDuration(duration)

			logger.Info("got incoming HTTP request",
				zap.String("method", r.Method),
				zap.String("URI", r.RequestURI),
				zap.Int("status", responseData.httpStatus),
				zap.Int("size", responseData.size),
				zap.String("duration", fmtDuration),
			)
		})
	}
}
