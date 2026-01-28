package processor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/handler/processor/mocks"
	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
)

type stubExpandShortener struct {
	retURL   string
	retErr   error
	retCount int
}

func (s *stubExpandShortener) IsReady() error {
	return nil
}

func (s *stubExpandShortener) Shorten(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
func (s *stubExpandShortener) Extract(_ context.Context, _ string) (string, error) {
	return s.retURL, s.retErr
}
func (s *stubExpandShortener) ShortenBatch(_ context.Context, _ string, _ []string) ([]string, error) {
	return nil, nil
}

func (s *stubExpandShortener) GetUserURLs(_ context.Context, _ string) ([]*model.URLStorageRecord, error) {
	return nil, nil
}

func (s *stubExpandShortener) DeleteBatch(_ context.Context, _ model.URLDeleteBatch) error {
	return nil
}

func (s *stubExpandShortener) Count(_ context.Context) (int, error) {
	return s.retCount, nil
}

func TestShortURLService_Expand(t *testing.T) {
	tests := []struct {
		name          string
		shortID       string
		stubOrigURL   string
		stubErr       error
		stubUrlsCount int
		wantOrigURL   string
		wantErr       bool
		wantErrIs     error
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
			shortener := &stubExpandShortener{tt.stubOrigURL, tt.stubErr, tt.stubUrlsCount}
			ep := mocks.NewMockAuditEventPublisher(t)
			if !tt.wantErr {
				ep.EXPECT().Publish(mock.AnythingOfType("model.AuditEvent")).Return().Once()
			}

			srv := NewExpand(shortener, zap.NewNop(), ep)
			ctx := auth.WithUser(context.Background(), &model.User{UUID: "userUUID"})

			gotURL, gotErr := srv.Process(ctx, tt.shortID)

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
