package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alex-storchak/shortener/internal/helper"
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
			fmtDuration := helper.FormatDuration(duration)

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

type gzipWriter struct {
	w   http.ResponseWriter
	gzw *gzip.Writer
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:   w,
		gzw: gzip.NewWriter(w),
	}
}

func (c *gzipWriter) Header() http.Header {
	return c.w.Header()
}

func (c *gzipWriter) Write(p []byte) (int, error) {
	return c.gzw.Write(p)
}

func (c *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *gzipWriter) Close() error {
	return c.gzw.Close()
}

type gzipReader struct {
	r   io.ReadCloser
	gzr *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:   r,
		gzr: gzr,
	}, nil
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.gzr.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.gzr.Close()
}

func canCompress(r *http.Request, compressEncType string) bool {
	var compressibleContentTypes = map[string]bool{
		`application/javascript`: true,
		`application/json`:       true,
		`text/css`:               true,
		`text/html`:              true,
		`text/plain`:             true,
		`text/xml`:               true,
	}

	canCompress := false
	contentTypeValues := r.Header.Values("Content-Type")
outer:
	for _, ct := range contentTypeValues {
		values := strings.Split(ct, ",")
		for _, v := range values {
			v = strings.Trim(v, " ")
			if _, ok := compressibleContentTypes[v]; ok {
				canCompress = true
				break outer
			}
		}
	}

	if !canCompress {
		return false
	}

	canCompress = false
	acceptEncodingValues := r.Header.Values("Accept-Encoding")
	for _, enc := range acceptEncodingValues {
		if strings.Contains(enc, compressEncType) {
			canCompress = true
		}
	}
	return canCompress
}

func isCompressed(r *http.Request, compressEncType string) bool {
	contentEncodingValues := r.Header.Values("Content-Encoding")
	for _, enc := range contentEncodingValues {
		if strings.Contains(enc, compressEncType) {
			return true
		}
	}
	return false
}

func GzipMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resWriter := w

			if canCompress(r, "gzip") {
				logger.Debug("Can compress response")
				gzw := newGzipWriter(w)
				resWriter = gzw
				defer gzw.Close()
			}

			if isCompressed(r, "gzip") {
				logger.Debug("Compressed data received")
				gzr, err := newGzipReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = gzr
				defer gzr.Close()
			}

			next.ServeHTTP(resWriter, r)
		})
	}
}
