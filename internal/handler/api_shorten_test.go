package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alex-storchak/shortener/internal/handler/mocks"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type ShortenSrvStub struct {
	shortenError error
}

func (s *ShortenSrvStub) Process(_ context.Context, _ model.ShortenRequest) (*model.ShortenResponse, string, error) {
	if s.shortenError != nil {
		return nil, "", s.shortenError
	}
	return &model.ShortenResponse{
		ShortURL: "https://example.com/abcde",
	}, "test", nil
}

func TestAPIShorten(t *testing.T) {
	type want struct {
		code        int
		body        model.ShortenResponse
		contentType string
	}
	tests := []struct {
		name         string
		method       string
		setupMock    func(m *mocks.MockAuditEventPublisher)
		want         want
		wantErr      bool
		shortenError error
		encodeError  error
	}{
		{
			name:   "POST request returns 201 (Created)",
			method: http.MethodPost,
			setupMock: func(m *mocks.MockAuditEventPublisher) {
				m.EXPECT().
					Publish(mock.Anything).
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
			name:      "returns 400 (Bad Request) when empty url in body requested",
			method:    http.MethodPost,
			setupMock: func(m *mocks.MockAuditEventPublisher) {},
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr:      true,
			shortenError: service.ErrEmptyInputURL,
		},
		// TODO: переписать на мок
		// {
		// 	name:      "returns 409 (Conflict) when url already exists",
		// 	method:    http.MethodPost,
		// 	setupMock: func(m *mocks.MockAuditEventPublisher) {},
		// 	want: want{
		// 		code: http.StatusConflict,
		// 		body: model.ShortenResponse{
		// 			ShortURL: "https://example.com/abcde",
		// 		},
		// 		contentType: "application/json",
		// 	},
		// 	wantErr:      false,
		// 	shortenError: service.ErrURLAlreadyExists,
		// },
		{
			name:      "returns 500 (Internal Server Error) when random error on shorten url",
			method:    http.MethodPost,
			setupMock: func(m *mocks.MockAuditEventPublisher) {},
			want: want{
				code: http.StatusInternalServerError,
			},
			wantErr:      true,
			shortenError: errors.New("random error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &ShortenSrvStub{tt.shortenError}

			mockPublisher := mocks.NewMockAuditEventPublisher(t)
			tt.setupMock(mockPublisher)

			h := handleAPIShorten(srv, zap.NewNop(), mockPublisher)

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
