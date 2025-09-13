package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alex-storchak/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type PingDBSrvStub struct {
	pingErr error
}

func (s *PingDBSrvStub) Ping() error {
	return s.pingErr
}

func TestPingDBHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		pingErr error
		want    int
	}{
		{
			name:   "non GET request returns 405 (Method Not Allowed)",
			method: http.MethodPost,
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "GET request returns 200 (OK) when DB ping succeeds",
			method: http.MethodGet,
			want:   http.StatusOK,
		},
		{
			name:    "GET request returns 500 (Internal Server Error) on ErrFailedToPingDB",
			method:  http.MethodGet,
			pingErr: service.ErrFailedToPingDB,
			want:    http.StatusInternalServerError,
		},
		{
			name:    "GET request returns 500 (Internal Server Error) on unknown error",
			method:  http.MethodGet,
			pingErr: errors.New("random error"),
			want:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &PingDBSrvStub{pingErr: tt.pingErr}
			h := NewPingDBHandler(srv, zap.NewNop())

			req := httptest.NewRequest(tt.method, "/ping", nil)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}
