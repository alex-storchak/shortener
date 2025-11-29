package processor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/handler/processor/mocks"
	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

type stubShortenerBatch struct {
	retIDs []string
	retErr error
}

func (s *stubShortenerBatch) IsReady() error {
	return nil
}

func (s *stubShortenerBatch) Shorten(_ context.Context, _ string, _ string) (string, error) {
	return "", nil
}

func (s *stubShortenerBatch) Extract(_ context.Context, _ string) (string, error) {
	return "", nil
}
func (s *stubShortenerBatch) ShortenBatch(_ context.Context, _ string, _ []string) ([]string, error) {
	return s.retIDs, s.retErr
}

func (s *stubShortenerBatch) GetUserURLs(_ context.Context, _ string) ([]*model.URLStorageRecord, error) {
	return nil, nil
}

func (s *stubShortenerBatch) DeleteBatch(_ context.Context, _ model.URLDeleteBatch) error {
	return nil
}

func TestShortenBatchService_ShortenBatch(t *testing.T) {
	tests := []struct {
		name         string
		decReq       model.BatchShortenRequest
		decErr       error
		stubShortIDs []string
		stubErr      error
		baseURL      string
		wantResp     model.BatchShortenResponse
		wantErr      bool
		wantErrIs    error
	}{
		{
			name:      "returns ErrEmptyInputBatch on empty request list",
			decReq:    []model.BatchShortenRequestItem{},
			stubErr:   service.ErrEmptyInputBatch,
			wantErr:   true,
			wantErrIs: service.ErrEmptyInputBatch,
		},
		{
			name: "returns ErrEmptyInputURL if any item has empty OriginalURL",
			decReq: []model.BatchShortenRequestItem{
				{CorrelationID: "1", OriginalURL: ""},
			},
			stubErr:   service.ErrEmptyInputURL,
			wantErr:   true,
			wantErrIs: service.ErrEmptyInputURL,
		},
		{
			name: "returns ErrEmptyInputURL from shortener",
			decReq: []model.BatchShortenRequestItem{
				{CorrelationID: "1", OriginalURL: "https://example.com"},
			},
			stubErr:   service.ErrEmptyInputURL,
			wantErr:   true,
			wantErrIs: service.ErrEmptyInputURL,
		},
		{
			name: "returns unexpected error",
			decReq: []model.BatchShortenRequestItem{
				{CorrelationID: "1", OriginalURL: "https://example.com"},
			},
			stubErr: errors.New("random error"),
			wantErr: true,
		},
		{
			name: "success builds response with baseURL and preserves order",
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
			shortener := &stubShortenerBatch{retIDs: tt.stubShortIDs, retErr: tt.stubErr}
			baseURL := tt.baseURL
			if baseURL == "" {
				baseURL = "http://any"
			}
			ub := mocks.NewMockShortURLBuilder(t)
			if !tt.wantErr {
				for _, s := range tt.stubShortIDs {
					ub.EXPECT().
						Build(s).
						Return(baseURL + "/" + s).
						Once()
				}
			}

			srv := NewAPIShortenBatch(shortener, zap.NewNop(), ub)
			ctx := auth.WithUser(context.Background(), &model.User{UUID: "userUUID"})

			resp, err := srv.Process(ctx, tt.decReq)

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
