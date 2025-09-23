package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stubShortener struct {
	retShortID string
	retErr     error
}

func (s *stubShortener) Shorten(_ string) (string, error) {
	return s.retShortID, s.retErr
}

func (s *stubShortener) Extract(_ string) (string, error) {
	return "", nil
}
func (s *stubShortener) ShortenBatch(_ *[]string) (*[]string, error) {
	return nil, nil
}

func TestMainPageService_Shorten(t *testing.T) {
	tests := []struct {
		name         string
		body         []byte
		stubShortID  string
		stubErr      error
		wantShortURL string
		wantErr      bool
		wantErrIs    error
	}{
		{
			name:      "returns ErrEmptyInputURL on empty body",
			body:      []byte(""),
			stubErr:   ErrEmptyInputURL,
			wantErr:   true,
			wantErrIs: ErrEmptyInputURL,
		},
		{
			name:         "returns shortURL on success",
			body:         []byte("https://example.com"),
			stubShortID:  "abcde",
			wantShortURL: "https://short.host/abcde",
			wantErr:      false,
		},
		{
			name:    "returns unexpected error",
			body:    []byte("https://example.com"),
			stubErr: errors.New("random error"),
			wantErr: true,
		},
		{
			name:         "returns shortURL and ErrURLAlreadyExists when URL bind exists in storage",
			body:         []byte("https://example.com"),
			stubShortID:  "exist",
			stubErr:      ErrURLAlreadyExists,
			wantShortURL: "https://short.host/exist",
			wantErr:      true,
			wantErrIs:    ErrURLAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := &stubShortener{tt.stubShortID, tt.stubErr}
			svc := NewMainPageService("https://short.host", core, zap.NewNop())

			gotURL, gotErr := svc.Shorten(tt.body)

			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, gotErr, tt.wantErrIs)
				}
				if tt.wantShortURL != "" {
					assert.Equal(t, tt.wantShortURL, gotURL)
				} else {
					assert.Equal(t, "", gotURL)
				}
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantShortURL, gotURL)
		})
	}
}
