package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/alex-storchak/shortener/internal/helper"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stubBatchDecoder struct {
	retReq []model.BatchShortenRequestItem
	retErr error
}

func (s *stubBatchDecoder) Decode(_ io.Reader) ([]model.BatchShortenRequestItem, error) {
	return s.retReq, s.retErr
}

type stubShortenerBatch struct {
	retIDs []string
	retErr error
}

func (s *stubShortenerBatch) IsReady() error {
	return nil
}

func (s *stubShortenerBatch) Shorten(_ string, _ string) (string, error) {
	return "", nil
}

func (s *stubShortenerBatch) Extract(_ string) (string, error) {
	return "", nil
}
func (s *stubShortenerBatch) ShortenBatch(_ string, _ []string) ([]string, error) {
	return s.retIDs, s.retErr
}

func (s *stubShortenerBatch) GetUserURLs(_ string) ([]*model.URLStorageRecord, error) {
	return nil, nil
}

func (s *stubShortenerBatch) DeleteBatch(_ model.URLDeleteBatch) error {
	return nil
}

func TestShortenBatchService_ShortenBatch(t *testing.T) {
	tests := []struct {
		name         string
		body         []byte
		decReq       []model.BatchShortenRequestItem
		decErr       error
		stubShortIDs []string
		stubErr      error
		baseURL      string
		wantResp     []model.BatchShortenResponseItem
		wantErr      bool
		wantErrIs    error
	}{
		{
			name:    "returns error when decoder fails",
			body:    []byte("[{bad json}]"),
			decErr:  errors.New("decode error"),
			wantErr: true,
		},
		{
			name:      "returns ErrEmptyInputBatch on empty request list",
			body:      []byte("[]"),
			decReq:    []model.BatchShortenRequestItem{},
			stubErr:   ErrEmptyInputBatch,
			wantErr:   true,
			wantErrIs: ErrEmptyInputBatch,
		},
		{
			name: "returns ErrEmptyInputURL if any item has empty OriginalURL",
			body: []byte("[{}]"),
			decReq: []model.BatchShortenRequestItem{
				{CorrelationID: "1", OriginalURL: ""},
			},
			stubErr:   ErrEmptyInputURL,
			wantErr:   true,
			wantErrIs: ErrEmptyInputURL,
		},
		{
			name: "returns ErrEmptyInputURL from shortener",
			body: []byte(`[{"correlation_id":"1","original_url":"https://example.com"}]}"`),
			decReq: []model.BatchShortenRequestItem{
				{CorrelationID: "1", OriginalURL: "https://example.com"},
			},
			stubErr:   ErrEmptyInputURL,
			wantErr:   true,
			wantErrIs: ErrEmptyInputURL,
		},
		{
			name: "returns unexpected error",
			body: []byte(`[{"correlation_id":"1","original_url":"https://example.com"}]`),
			decReq: []model.BatchShortenRequestItem{
				{CorrelationID: "1", OriginalURL: "https://example.com"},
			},
			stubErr: errors.New("random error"),
			wantErr: true,
		},
		{
			name: "success builds response with baseURL and preserves order",
			body: []byte("[]"),
			decReq: []model.BatchShortenRequestItem{
				{CorrelationID: "1", OriginalURL: "https://a"},
				{CorrelationID: "2", OriginalURL: "https://b"},
			},
			stubShortIDs: []string{"abc", "def"},
			baseURL:      "http://short.host",
			wantResp: []model.BatchShortenResponseItem{
				{CorrelationID: "1", ShortURL: "http://short.host/abc"},
				{CorrelationID: "2", ShortURL: "http://short.host/def"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := &stubBatchDecoder{retReq: tt.decReq, retErr: tt.decErr}
			shortener := &stubShortenerBatch{retIDs: tt.stubShortIDs, retErr: tt.stubErr}
			baseURL := tt.baseURL
			if baseURL == "" {
				baseURL = "http://any"
			}
			srv := NewShortenBatchService(baseURL, shortener, dec, zap.NewNop())
			ctx := helper.WithUser(context.Background(), &model.User{UUID: "userUUID"})

			var r io.Reader = bytes.NewReader(tt.body)
			resp, err := srv.Process(ctx, r)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				}
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantResp, resp)
		})
	}
}
