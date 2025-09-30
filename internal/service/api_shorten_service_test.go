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

type stubDecoder struct {
	retReq model.ShortenRequest
	retErr error
}

func (s *stubDecoder) Decode(_ io.Reader) (model.ShortenRequest, error) {
	return s.retReq, s.retErr
}

type stubShortenerAPI struct {
	retShortID string
	retErr     error
}

func (s *stubShortenerAPI) Shorten(_, _ string) (string, error) {
	return s.retShortID, s.retErr
}

func (s *stubShortenerAPI) Extract(_ string) (string, error) {
	return "", nil
}

func (s *stubShortenerAPI) ShortenBatch(_ string, _ *[]string) (*[]string, error) {
	return nil, nil
}

func (s *stubShortenerAPI) GetUserURLs(_ string) (*[]model.URLStorageRecord, error) {
	return nil, nil
}

func TestAPIShortenService_Shorten(t *testing.T) {
	tests := []struct {
		name       string
		body       []byte
		decoderReq model.ShortenRequest
		decoderErr error
		shortID    string
		shortenErr error
		baseURL    string
		wantResp   *model.ShortenResponse
		wantErr    bool
		wantErrIs  error
	}{
		{
			name:       "returns error when decoder fails",
			body:       []byte("{bad json}"),
			decoderErr: errors.New("decode error"),
			wantErr:    true,
		},
		{
			name:       "returns ErrEmptyInputURL from shortener",
			body:       []byte("{}"),
			decoderReq: model.ShortenRequest{OrigURL: ""},
			shortenErr: ErrEmptyInputURL,
			wantErr:    true,
			wantErrIs:  ErrEmptyInputURL,
		},
		{
			name:       "returns unexpected shortener error",
			body:       []byte("{}"),
			decoderReq: model.ShortenRequest{OrigURL: "https://example.com"},
			shortenErr: errors.New("random error"),
			wantErr:    true,
		},
		{
			name:       "success returns short url in response",
			body:       []byte("{}"),
			decoderReq: model.ShortenRequest{OrigURL: "https://example.com"},
			shortID:    "abcde",
			baseURL:    "https://short.host",
			wantResp:   &model.ShortenResponse{ShortURL: "https://short.host/abcde"},
		},
		{
			name:       "returns short url and ErrURLAlreadyExists when URL bind exists in storage",
			body:       []byte("{}"),
			decoderReq: model.ShortenRequest{OrigURL: "https://example.com"},
			shortID:    "exist",
			shortenErr: ErrURLAlreadyExists,
			baseURL:    "https://short.host",
			wantResp:   &model.ShortenResponse{ShortURL: "https://short.host/exist"},
			wantErr:    true,
			wantErrIs:  ErrURLAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := &stubDecoder{tt.decoderReq, tt.decoderErr}
			shortener := &stubShortenerAPI{tt.shortID, tt.shortenErr}
			baseURL := tt.baseURL
			if baseURL == "" {
				baseURL = "http://any"
			}
			srv := NewAPIShortenService(baseURL, shortener, dec, zap.NewNop())
			ctx := context.WithValue(context.Background(), helper.UserCtxKey{}, &model.User{UUID: "userUUID"})

			var r io.Reader = bytes.NewReader(tt.body)
			resp, err := srv.Shorten(ctx, r)

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
