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
	return "", errors.New("")
}

func (s *stubShortener) ShortenBatch(_ *[]string) (*[]string, error) {
	return nil, nil
}

func TestShortenCore_Shorten(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		OrigURL      string
		stubShortID  string
		stubErr      error
		wantShortURL string
		wantShortID  string
		wantErr      bool
		wantErrIs    error
	}{
		{
			name:      "returns error ErrEmptyInputURL on empty input",
			baseURL:   "https://example.com",
			OrigURL:   "",
			wantErr:   true,
			wantErrIs: ErrEmptyInputURL,
		},
		{
			name:         "returns composed short URL and id on success",
			baseURL:      "https://short.host",
			OrigURL:      "https://example.com/",
			stubShortID:  "abcde",
			wantShortID:  "abcde",
			wantShortURL: "https://short.host/abcde",
			wantErr:      false,
		},
		{
			name:    "returns error on failed shortener",
			baseURL: "https://short.host",
			OrigURL: "https://example.com/",
			stubErr: errors.New("failed to shorten"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortenerStub := &stubShortener{tt.stubShortID, tt.stubErr}
			core := NewShortenCore(shortenerStub, tt.baseURL, zap.NewNop())

			gotShortURL, gotShortID, gotErr := core.Shorten(tt.OrigURL)

			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, gotErr, tt.wantErrIs)
				}
				assert.Equal(t, "", gotShortURL)
				assert.Equal(t, "", gotShortID)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantShortURL, gotShortURL)
			assert.Equal(t, tt.wantShortID, gotShortID)
		})
	}
}
