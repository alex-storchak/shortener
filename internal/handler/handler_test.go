package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alex-storchak/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type shortenerStub struct {
	urlStorage      map[string]string
	shortURLStorage map[string]string
}

func newShortenerStub() *shortenerStub {
	return &shortenerStub{
		urlStorage: map[string]string{},
		shortURLStorage: map[string]string{
			"abcde": "https://existing.com",
		},
	}
}

func (d *shortenerStub) Shorten(url string) (string, error) {
	return "abcde", nil
}

func (d *shortenerStub) Extract(shortID string) (string, error) {
	if targetURL, ok := d.shortURLStorage[shortID]; ok {
		return targetURL, nil
	} else {
		return "", service.ErrShortenerShortIDNotFound
	}
}

func Test_handlers_MainPageHandler(t *testing.T) {
	type want struct {
		code        int
		body        string
		contentType string
	}
	tests := []struct {
		name    string
		method  string
		want    want
		wantErr bool
	}{
		{
			name:   "non POST request returns 405 (Method Not Allowed)",
			method: http.MethodGet,
			want: want{
				code: http.StatusMethodNotAllowed,
			},
			wantErr: true,
		},
		{
			name:   "POST request returns 201 (Created)",
			method: http.MethodPost,
			want: want{
				code:        http.StatusCreated,
				body:        "http://example.com/abcde",
				contentType: "text/plain",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				shortener:        newShortenerStub(),
				shortURLBaseAddr: "http://example.com",
			}
			request := httptest.NewRequest(tt.method, "/", strings.NewReader("http://existing.com"))
			w := httptest.NewRecorder()

			h.MainPageHandler(w, request)
			res := w.Result()

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return
			}

			if tt.wantErr {
				assert.Equal(t, tt.want.code, res.StatusCode)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.body, string(resBody))
		})
	}
}

func Test_handlers_ShortURLHandler(t *testing.T) {
	type want struct {
		code     int
		Location string
	}
	tests := []struct {
		name    string
		method  string
		path    string
		shortID string
		want    want
		wantErr bool
	}{
		{
			name:   "non GET request returns 405 (Method Not Allowed)",
			method: http.MethodPost,
			path:   "/abcde",
			want: want{
				code: http.StatusMethodNotAllowed,
			},
			wantErr: true,
		},
		{
			name:    "existing short url returns 307 (Temporary Redirect)",
			method:  http.MethodGet,
			path:    "/abcde",
			shortID: "abcde",
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
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				shortener: newShortenerStub(),
			}
			request := httptest.NewRequest(tt.method, tt.path, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.shortID)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			h.ShortURLHandler(w, request)
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
