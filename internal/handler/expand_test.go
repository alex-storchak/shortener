package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	repo "github.com/alex-storchak/shortener/internal/repository"
)

type ShortURLSrvStub struct {
	expandError error
}

func (s *ShortURLSrvStub) Process(_ context.Context, _ string) (origURL string, err error) {
	if s.expandError != nil {
		return "", s.expandError
	}
	return "https://existing.com", nil
}

func TestExpand(t *testing.T) {
	type want struct {
		code     int
		Location string
	}
	tests := []struct {
		name        string
		method      string
		path        string
		want        want
		wantErr     bool
		expandError error
	}{
		{
			name:   "existing short url returns 307 (Temporary Redirect)",
			method: http.MethodGet,
			path:   "/abcde",
			want: want{
				code:     http.StatusTemporaryRedirect,
				Location: "https://existing.com",
			},
			wantErr: false,
		},
		{
			name:   "non-existing short url returns 404 (Not Found)",
			method: http.MethodGet,
			path:   "/non-existing",
			want: want{
				code: http.StatusNotFound,
			},
			wantErr:     true,
			expandError: repo.NewDataNotFoundError(nil),
		},
		{
			name:   "returns 500 (Internal Server Error) when random error on expand happens",
			method: http.MethodGet,
			path:   "/non-existing",
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

			h := HandleExpand(srv, zap.NewNop())

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
