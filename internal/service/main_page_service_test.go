package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stubShortenCore struct {
	retShortURL string
	retErr      error
}

func (s *stubShortenCore) Shorten(_ string) (string, string, error) {
	return s.retShortURL, "", s.retErr
}

func TestMainPageService_Shorten(t *testing.T) {
	tests := []struct {
		name         string
		body         []byte
		stubShortURL string
		stubErr      error
		wantShortURL string
		wantErr      bool
		wantErrIs    error
	}{
		{
			name:      "maps ErrEmptyInputURL to ErrEmptyBody on empty body",
			body:      []byte(""),
			stubErr:   ErrEmptyInputURL,
			wantErr:   true,
			wantErrIs: ErrEmptyBody,
		},
		{
			name:         "returns shortURL on success",
			body:         []byte("https://example.com"),
			stubShortURL: "https://short.host/abcde",
			wantShortURL: "https://short.host/abcde",
			wantErr:      false,
		},
		{
			name:    "returns unexpected error",
			body:    []byte("https://example.com"),
			stubErr: errors.New("random error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := &stubShortenCore{tt.stubShortURL, tt.stubErr}
			svc := NewMainPageService(core, zap.NewNop())

			gotURL, gotErr := svc.Shorten(tt.body)

			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, gotErr, tt.wantErrIs)
				}
				assert.Equal(t, "", gotURL)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantShortURL, gotURL)
		})
	}
}
