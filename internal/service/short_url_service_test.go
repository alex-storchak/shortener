package service

import (
	"errors"
	"testing"

	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stubExpandShortener struct {
	retURL string
	retErr error
}

func (s *stubExpandShortener) Shorten(_ string) (string, error) {
	return "", nil
}
func (s *stubExpandShortener) Extract(_ string) (string, error) {
	return s.retURL, s.retErr
}
func (s *stubExpandShortener) ShortenBatch(_ *[]string) (*[]string, error) {
	return nil, nil
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
			name:      "maps ErrURLStorageDataNotFound to ErrShortURLNotFound",
			shortID:   "abcde",
			stubErr:   repository.ErrURLStorageDataNotFound,
			wantErr:   true,
			wantErrIs: ErrShortURLNotFound,
		},
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
			svc := NewShortURLService(shortener, zap.NewNop())

			gotURL, gotErr := svc.Expand(tt.shortID)

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
