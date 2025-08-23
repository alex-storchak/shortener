package logger

import (
	"net/http"
	"time"

	"github.com/alex-storchak/shortener/internal/logger/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetInstance(cfg *config.Config) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}
	zcfg := zap.NewProductionConfig()
	zcfg.Level = lvl
	zcfg.Encoding = "console"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zl, err := zcfg.Build()
	if err != nil {
		return nil, err
	}

	return zl, nil
}

func RequestLogger(h http.HandlerFunc, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		h(&lw, r)

		duration := time.Since(start)
		logger.Info("got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("URI", r.RequestURI),
			zap.Int("status", responseData.httpStatus),
			zap.Int("size", responseData.size),
			zap.Duration("duration", duration),
		)
	}
}

type responseData struct {
	httpStatus int
	size       int
}

// добавляем реализацию http.ResponseWriter
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
