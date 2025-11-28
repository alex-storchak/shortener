package processor

import (
	"context"
	"errors"
	"testing"

	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stubShortenerAPI struct {
	retShortID string
	retErr     error
}

func (s *stubShortenerAPI) Shorten(_ context.Context, _, _ string) (string, error) {
	return s.retShortID, s.retErr
}

func (s *stubShortenerAPI) Extract(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (s *stubShortenerAPI) ShortenBatch(_ context.Context, _ string, _ []string) ([]string, error) {
	return nil, nil
}

func (s *stubShortenerAPI) GetUserURLs(_ context.Context, _ string) ([]*model.URLStorageRecord, error) {
	return nil, nil
}

func (s *stubShortenerAPI) DeleteBatch(_ context.Context, _ model.URLDeleteBatch) error {
	return nil
}

func TestShortenService_Shorten(t *testing.T) {
	tests := []struct {
		name       string
		decReq     model.ShortenRequest
		decoderErr error
		shortID    string
		shortenErr error
		baseURL    string
		wantResp   *model.ShortenResponse
		wantErr    bool
		wantErrIs  error
	}{
		{
			name:       "returns ErrEmptyInputURL from shortener",
			decReq:     model.ShortenRequest{OrigURL: ""},
			shortenErr: service.ErrEmptyInputURL,
			wantErr:    true,
			wantErrIs:  service.ErrEmptyInputURL,
		},
		{
			name:       "returns unexpected shortener error",
			decReq:     model.ShortenRequest{OrigURL: "https://example.com"},
			shortenErr: errors.New("random error"),
			wantErr:    true,
		},
		{
			name:     "success returns short url in response",
			decReq:   model.ShortenRequest{OrigURL: "https://example.com"},
			shortID:  "abcde",
			baseURL:  "https://short.host",
			wantResp: &model.ShortenResponse{ShortURL: "https://short.host/abcde"},
		},
		{
			name:       "returns short url and ErrURLAlreadyExists when URL bind exists in storage",
			decReq:     model.ShortenRequest{OrigURL: "https://example.com"},
			shortID:    "exist",
			shortenErr: service.ErrURLAlreadyExists,
			baseURL:    "https://short.host",
			wantResp:   &model.ShortenResponse{ShortURL: "https://short.host/exist"},
			wantErr:    true,
			wantErrIs:  service.ErrURLAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortener := &stubShortenerAPI{tt.shortID, tt.shortenErr}
			baseURL := tt.baseURL
			if baseURL == "" {
				baseURL = "http://any"
			}
			srv := NewAPIShorten(baseURL, shortener, zap.NewNop())
			ctx := auth.WithUser(context.Background(), &model.User{UUID: "userUUID"})

			resp, _, err := srv.Process(ctx, tt.decReq)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				}
				if tt.wantResp != nil {
					assert.Equal(t, tt.wantResp, resp)
				} else {
					assert.Nil(t, resp)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantResp, resp)
		})
	}
}
