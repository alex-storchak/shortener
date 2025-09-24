package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alex-storchak/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MainPageSrvStub struct {
	shortenError error
}

func (s *MainPageSrvStub) Shorten(_ []byte) (shortURL string, err error) {
	if s.shortenError != nil {
		return "", s.shortenError
	}
	return "https://example.com/abcde", nil
}

func TestMainPageHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code        int
		body        string
		contentType string
	}
	tests := []struct {
		name         string
		method       string
		want         want
		wantErr      bool
		shortenError error
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
				body:        "https://example.com/abcde",
				contentType: "text/plain",
			},
			wantErr: false,
		},
		{
			name:   "POST request returns 400 (Bad Request) when shorten errors on empty body passed",
			method: http.MethodPost,
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr:      true,
			shortenError: service.ErrEmptyBody,
		},
		{
			name:   "POST request returns 500 (Internal Server Error) when random error on shorten happens",
			method: http.MethodPost,
			want: want{
				code: http.StatusInternalServerError,
			},
			wantErr:      true,
			shortenError: errors.New("random error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &MainPageSrvStub{tt.shortenError}
			h := NewMainPageHandler(srv, zap.NewNop())

			request := httptest.NewRequest(tt.method, "/", strings.NewReader("https://existing.com"))
			w := httptest.NewRecorder()

			h.ServeHTTP(w, request)
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
