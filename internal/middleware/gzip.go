package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

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
		return nil, fmt.Errorf("create gzip reader: %w", err)
	}

	return &gzipReader{
		r:   r,
		gzr: gzr,
	}, nil
}

func (c *gzipReader) Read(p []byte) (n int, err error) {
	return c.gzr.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("close gzip reader: %w", err)
	}
	return c.gzr.Close()
}

func hasCompressibleContentType(r *http.Request) bool {
	var compressibleContentTypes = map[string]struct{}{
		`application/javascript`: {},
		`application/json`:       {},
		`text/css`:               {},
		`text/html`:              {},
		`text/plain`:             {},
		`text/xml`:               {},
	}

	contentTypeValues := r.Header.Values("Content-Type")
	for _, ct := range contentTypeValues {
		values := strings.Split(ct, ",")
		for _, v := range values {
			v = strings.TrimSpace(v)
			if _, ok := compressibleContentTypes[v]; ok {
				return true
			}
		}
	}
	return false
}

func isAcceptsEncoding(r *http.Request, compressEncType string) bool {
	acceptEncodingValues := r.Header.Values("Accept-Encoding")
	for _, enc := range acceptEncodingValues {
		if strings.Contains(enc, compressEncType) {
			return true
		}
	}
	return false
}

func canCompress(r *http.Request, compressEncType string) bool {
	if !hasCompressibleContentType(r) {
		return false
	}
	return isAcceptsEncoding(r, compressEncType)
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

func wrapWriterWithCompress(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, io.Closer) {
	if canCompress(r, "gzip") {
		gzw := newGzipWriter(w)
		return gzw, gzw
	}
	return w, nil
}

func wrapReqBodyWithDecompress(r *http.Request) (io.Closer, error) {
	if isCompressed(r, "gzip") {
		gzr, err := newGzipReader(r.Body)
		if err != nil {
			return nil, fmt.Errorf("create new gzip reader: %w", err)
		}
		r.Body = gzr
		return gzr, nil
	}
	return nil, nil
}

func NewGzip(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resWriter, respCloser := wrapWriterWithCompress(w, r)
			if respCloser != nil {
				defer respCloser.Close()
			}

			reqCloser, err := wrapReqBodyWithDecompress(r)
			if err != nil {
				logger.Error("failed to wrap request body with decompress", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if reqCloser != nil {
				defer reqCloser.Close()
			}

			next.ServeHTTP(resWriter, r)
		})
	}
}
