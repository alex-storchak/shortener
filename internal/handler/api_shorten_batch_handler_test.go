package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type APIShortenBatchSrvStub struct {
	shortenError error
}

func (s *APIShortenBatchSrvStub) ShortenBatch(_ context.Context, _ io.Reader) ([]model.BatchShortenResponseItem, error) {
	if s.shortenError != nil {
		return nil, s.shortenError
	}
	return []model.BatchShortenResponseItem{
		{CorrelationID: "1", ShortURL: "https://example.com/a1"},
		{CorrelationID: "2", ShortURL: "https://example.com/b2"},
	}, nil
}

type JSONEncoderBatchStub struct {
	encodeError error
}

func (e *JSONEncoderBatchStub) Encode(w io.Writer, _ any) error {
	if e.encodeError != nil {
		return e.encodeError
	}
	_, err := w.Write([]byte(`[{"correlation_id":"1","short_url":"https://example.com/a1"},{"correlation_id":"2","short_url":"https://example.com/b2"}]`))
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}

func TestAPIShortenBatchHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code        int
		body        []model.BatchShortenResponseItem
		contentType string
	}
	tests := []struct {
		name         string
		method       string
		contentType  string
		want         want
		wantErr      bool
		shortenError error
		encodeError  error
	}{
		{
			name:        "wrong Content-Type returns 400 (Bad Request)",
			method:      http.MethodPost,
			contentType: "text/plain",
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr: true,
		},
		{
			name:        "POST request returns 201 (Created)",
			method:      http.MethodPost,
			contentType: "application/json",
			want: want{
				code: http.StatusCreated,
				body: []model.BatchShortenResponseItem{
					{CorrelationID: "1", ShortURL: "https://example.com/a1"},
					{CorrelationID: "2", ShortURL: "https://example.com/b2"},
				},
				contentType: "application/json",
			},
			wantErr: false,
		},
		{
			name:        "returns 400 (Bad Request) when empty batch provided",
			method:      http.MethodPost,
			contentType: "application/json",
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr:      true,
			shortenError: service.ErrEmptyInputBatch,
		},
		{
			name:        "returns 400 (Bad Request) when empty url in request json",
			method:      http.MethodPost,
			contentType: "application/json",
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr:      true,
			shortenError: service.ErrEmptyInputURL,
		},
		{
			name:        "returns 500 (Internal Server Error) when random error on shorten batch",
			method:      http.MethodPost,
			contentType: "application/json",
			want: want{
				code: http.StatusInternalServerError,
			},
			wantErr:      true,
			shortenError: errors.New("random error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			srv := &APIShortenBatchSrvStub{tt.shortenError}
			enc := &JSONEncoderBatchStub{tt.encodeError}
			h := NewAPIShortenBatchHandler(srv, enc, zap.NewNop())

			request := httptest.NewRequest(tt.method, "/api/shorten/batch", strings.NewReader(`[{"correlation_id":"1","original_url":"https://existing.com/1"}]`))
			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			if tt.wantErr {
				assert.Equal(t, tt.want.code, res.StatusCode)
				return
			}

			var respItems []model.BatchShortenResponseItem
			err := json.NewDecoder(res.Body).Decode(&respItems)

			require.NoError(t, err)
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.body, respItems)
		})
	}
}
