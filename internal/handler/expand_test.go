package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alex-storchak/shortener/internal/handler/mocks"
	repo "github.com/alex-storchak/shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type ShortURLSrvStub struct {
	expandError error
}

func (s *ShortURLSrvStub) Process(_ context.Context, _ string) (origURL, userUUID string, err error) {
	if s.expandError != nil {
		return "", "userUUID", s.expandError
	}
	return "https://existing.com", "userUUID", nil
}

func TestExpand(t *testing.T) {
	type want struct {
		code     int
		Location string
	}
	tests := []struct {
		name        string
		method      string
		setupMock   func(m *mocks.MockAuditEventPublisher)
		path        string
		want        want
		wantErr     bool
		expandError error
	}{
		{
			name:   "existing short url returns 307 (Temporary Redirect)",
			method: http.MethodGet,
			setupMock: func(m *mocks.MockAuditEventPublisher) {
				m.EXPECT().
					Publish(mock.AnythingOfType("model.AuditEvent")).
					Once()
			},
			path: "/abcde",
			want: want{
				code:     http.StatusTemporaryRedirect,
				Location: "https://existing.com",
			},
			wantErr: false,
		},
		{
			name:      "non-existing short url returns 404 (Not Found)",
			method:    http.MethodGet,
			setupMock: func(m *mocks.MockAuditEventPublisher) {},
			path:      "/non-existing",
			want: want{
				code: http.StatusNotFound,
			},
			wantErr:     true,
			expandError: repo.NewDataNotFoundError(nil),
		},
		{
			name:      "returns 500 (Internal Server Error) when random error on expand happens",
			method:    http.MethodGet,
			setupMock: func(m *mocks.MockAuditEventPublisher) {},
			path:      "/non-existing",
			want: want{
				code: http.StatusInternalServerError,
			},
			wantErr:     true,
			expandError: errors.New("random error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &ShortURLSrvStub{tt.expandError}

			mockPublisher := mocks.NewMockAuditEventPublisher(t)
			tt.setupMock(mockPublisher)

			h := handleExpand(srv, zap.NewNop(), mockPublisher)

			request := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			if tt.wantErr {
				assert.Equal(t, tt.want.code, res.StatusCode)
				return
			}

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.Location, res.Header.Get("Location"))
		})
	}
}
