package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// gzipWriter wraps http.ResponseWriter with gzip compression.
type gzipWriter struct {
	w   http.ResponseWriter
	gzw *gzip.Writer
}

// newGzipWriter creates a new gzip writer wrapper.
func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:   w,
		gzw: gzip.NewWriter(w),
	}
}

// Header returns the header map from the underlying ResponseWriter.
func (c *gzipWriter) Header() http.Header {
	return c.w.Header()
}

// Write compresses the data and writes it to the underlying gzip writer.
func (c *gzipWriter) Write(p []byte) (int, error) {
	return c.gzw.Write(p)
}

// WriteHeader sets the status code and adds Content-Encoding header for successful responses.
func (c *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes the gzip writer and flushes any pending data.
func (c *gzipWriter) Close() error {
	return c.gzw.Close()
}

// gzipReader wraps io.ReadCloser with gzip decompression.
type gzipReader struct {
	r   io.ReadCloser
	gzr *gzip.Reader
}

// newGzipReader creates a new gzip reader wrapper.
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

// Read decompresses data from the underlying gzip reader.
func (c *gzipReader) Read(p []byte) (n int, err error) {
	return c.gzr.Read(p)
}

// Close closes both the original reader and the gzip reader.
func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("close gzip reader: %w", err)
	}
	return c.gzr.Close()
}

// hasCompressibleContentType checks if the request content type is compressible.
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

// isAcceptsEncoding checks if the client accepts the specified compression encoding.
func isAcceptsEncoding(r *http.Request, compressEncType string) bool {
	acceptEncodingValues := r.Header.Values("Accept-Encoding")
	for _, enc := range acceptEncodingValues {
		if strings.Contains(enc, compressEncType) {
			return true
		}
	}
	return false
}

// canCompress determines if the response should be compressed based on content type and client acceptance.
func canCompress(r *http.Request, compressEncType string) bool {
	if !hasCompressibleContentType(r) {
		return false
	}
	return isAcceptsEncoding(r, compressEncType)
}

// isCompressed checks if the request body is already compressed with the specified encoding.
func isCompressed(r *http.Request, compressEncType string) bool {
	contentEncodingValues := r.Header.Values("Content-Encoding")
	for _, enc := range contentEncodingValues {
		if strings.Contains(enc, compressEncType) {
			return true
		}
	}
	return false
}

// wrapWriterWithCompress wraps the response writer with gzip compression if applicable.
func wrapWriterWithCompress(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, io.Closer) {
	if canCompress(r, "gzip") {
		gzw := newGzipWriter(w)
		return gzw, gzw
	}
	return w, nil
}

// wrapReqBodyWithDecompress wraps the request body with gzip decompression if applicable.
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

// NewGzip creates middleware that handles gzip compression and decompression.
// The middleware provides:
//   - Automatic compression of responses for compressible content types
//   - Automatic decompression of gzip-compressed request bodies
//   - Content type filtering for compression
//   - Client capability detection via Accept-Encoding header
//
// Parameters:
//   - logger: structured logger for logging operations
//
// Returns:
//   - func(http.Handler) http.Handler: gzip compression middleware function
//
// The middleware handles both request decompression and response compression
// transparently, reducing bandwidth usage without requiring handler changes.
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
