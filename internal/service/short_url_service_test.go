package service

import (
	"errors"
	"testing"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stubExpandShortener struct {
	retURL string
	retErr error
}

func (s *stubExpandShortener) IsReady() error {
	return nil
}

func (s *stubExpandShortener) Shorten(_, _ string) (string, error) {
	return "", nil
}
func (s *stubExpandShortener) Extract(_ string) (string, error) {
	return s.retURL, s.retErr
}
func (s *stubExpandShortener) ShortenBatch(_ string, _ []string) ([]string, error) {
	return nil, nil
}

func (s *stubExpandShortener) GetUserURLs(_ string) ([]*model.URLStorageRecord, error) {
	return nil, nil
}

func (s *stubExpandShortener) DeleteBatch(_ model.URLDeleteBatch) error {
	return nil
}

func TestShortURLService_Expand(t *testing.T) {
	tests := []struct {
		name        string
		shortID     string
		stubOrigURL string
		stubErr     error
		wantOrigURL string
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:        "returns original url on success",
			shortID:     "abcde",
			stubOrigURL: "https://example.com",
			wantOrigURL: "https://example.com",
			wantErr:     false,
		},
		{
			name:    "returns unexpected error",
			shortID: "abcde",
			stubErr: errors.New("random error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortener := &stubExpandShortener{tt.stubOrigURL, tt.stubErr}
			srv := NewShortURLService(shortener, zap.NewNop())

			gotURL, gotErr := srv.Process(tt.shortID)

			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, gotErr, tt.wantErrIs)
				}
				assert.Equal(t, "", gotURL)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantOrigURL, gotURL)
		})
	}
}
