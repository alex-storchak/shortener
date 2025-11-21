package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGzipMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		acceptEncoding   string
		expectedEncoding string
		contentEncoding  string
		shouldCompress   bool
		body             string
		contentType      string
	}{
		{
			name:             "returns compressed data when client supports gzip and contentType supported",
			acceptEncoding:   "gzip, deflate",
			expectedEncoding: "gzip",
			shouldCompress:   true,
			body:             "http://www.example.com",
			contentType:      "text/plain",
		},
		{
			name:             "returns non-compressed data when client does not support gzip",
			acceptEncoding:   "",
			expectedEncoding: "",
			shouldCompress:   false,
			body:             "http://www.example.com",
			contentType:      "text/plain",
		},
		{
			name:             "returns non-compressed data when client supports gzip but contentType doesn't supported",
			acceptEncoding:   "gzip, deflate",
			expectedEncoding: "",
			shouldCompress:   false,
			body:             "fake image",
			contentType:      "image/jpeg",
		},
		{
			name:             "returns compressed data when client supports gzip contentType and compressed data sent",
			acceptEncoding:   "gzip, deflate",
			expectedEncoding: "gzip",
			contentEncoding:  "gzip",
			shouldCompress:   true,
			body:             "http://www.example.com",
			contentType:      "text/plain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler(tt.contentType)
			gzippedHandler := GzipMiddleware(zap.NewNop())(handler)

			buf, err := makeRequestBody(tt.body, tt.contentEncoding)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/", buf)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			req.Header.Set("Content-Type", tt.contentType)
			req.Header.Set("Content-Encoding", tt.contentEncoding)
			rr := httptest.NewRecorder()

			gzippedHandler.ServeHTTP(rr, req)

			encoding := rr.Header().Get("Content-Encoding")
			assert.Equal(t, tt.expectedEncoding, encoding)
			if tt.shouldCompress {
				assert.True(t, isValidGzip(rr.Body.Bytes()))
				decompressed, err := decompressGzip(rr.Body.Bytes())
				require.NoError(t, err)
				assert.Equal(t, tt.body, string(decompressed))
			} else {
				assert.Equal(t, tt.body, rr.Body.String())
			}
		})
	}
}

func newTestHandler(contentType string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})
}

func makeRequestBody(body string, contentEncoding string) (*bytes.Buffer, error) {
	if contentEncoding == "gzip" {
		buf := bytes.NewBuffer(nil)
		zw := gzip.NewWriter(buf)
		if _, err := zw.Write([]byte(body)); err != nil {
			return nil, fmt.Errorf("write req body (%s) with gzip writer: %w", body, err)
		}
		if err := zw.Close(); err != nil {
			return nil, fmt.Errorf("close gzip writer: %w", err)
		}
		return buf, nil
	}
	return bytes.NewBufferString(body), nil
}

func isValidGzip(data []byte) bool {
	_, err := gzip.NewReader(bytes.NewReader(data))
	return err == nil
}

func decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create gzip reader: %w", err)
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
