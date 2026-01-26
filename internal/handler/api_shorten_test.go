package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/handler/mocks"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

func TestAPIShorten(t *testing.T) {
	type want struct {
		code        int
		body        model.ShortenResponse
		contentType string
	}
	tests := []struct {
		name          string
		method        string
		setupMockProc func(m *mocks.MockAPIShortenProcessor)
		want          want
		wantErr       bool
	}{
		{
			name:   "POST request returns 201 (Created)",
			method: http.MethodPost,
			setupMockProc: func(m *mocks.MockAPIShortenProcessor) {
				m.EXPECT().
					Process(mock.Anything, mock.AnythingOfType("model.ShortenRequest")).
					Return(&model.ShortenResponse{ShortURL: "https://example.com/abcde"}, nil).
					Once()
			},
			want: want{
				code: http.StatusCreated,
				body: model.ShortenResponse{
					ShortURL: "https://example.com/abcde",
				},
				contentType: "application/json",
			},
			wantErr: false,
		},
		{
			name:   "returns 400 (Bad Request) when empty url in body requested",
			method: http.MethodPost,
			setupMockProc: func(m *mocks.MockAPIShortenProcessor) {
				m.EXPECT().
					Process(mock.Anything, mock.AnythingOfType("model.ShortenRequest")).
					Return(&model.ShortenResponse{}, service.ErrEmptyInputURL).
					Once()
			},
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr: true,
		},
		{
			name:   "returns 409 (Conflict) when url already exists",
			method: http.MethodPost,
			setupMockProc: func(m *mocks.MockAPIShortenProcessor) {
				m.EXPECT().
					Process(mock.Anything, mock.AnythingOfType("model.ShortenRequest")).
					Return(&model.ShortenResponse{ShortURL: "https://example.com/abcde"}, service.ErrURLAlreadyExists).
					Once()
			},
			want: want{
				code: http.StatusConflict,
				body: model.ShortenResponse{
					ShortURL: "https://example.com/abcde",
				},
				contentType: "application/json",
			},
			wantErr: false,
		},
		{
			name:   "returns 500 (Internal Server Error) when random error on shorten url",
			method: http.MethodPost,
			setupMockProc: func(m *mocks.MockAPIShortenProcessor) {
				m.EXPECT().
					Process(mock.Anything, mock.AnythingOfType("model.ShortenRequest")).
					Return(&model.ShortenResponse{}, errors.New("random error")).
					Once()
			},
			want: want{
				code: http.StatusInternalServerError,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProc := mocks.NewMockAPIShortenProcessor(t)
			tt.setupMockProc(mockProc)

			h := HandleAPIShorten(mockProc, zap.NewNop())

			request := httptest.NewRequest(tt.method, "/", strings.NewReader(`{"url":"https://existing.com"}`))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			if tt.wantErr {
				assert.Equal(t, tt.want.code, res.StatusCode)
				return
			}

			var shortenResponse model.ShortenResponse
			err := json.NewDecoder(res.Body).Decode(&shortenResponse)

			require.NoError(t, err)
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.body, shortenResponse)
		})
	}
}
