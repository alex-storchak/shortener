package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type PingSrvStub struct {
	pingErr error
}

func (s *PingSrvStub) Process() error {
	return s.pingErr
}

func TestPingHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		pingErr error
		want    int
	}{
		{
			name:   "GET request returns 200 (OK) when DB ping succeeds",
			method: http.MethodGet,
			want:   http.StatusOK,
		},
		{
			name:    "GET request returns 500 (Internal Server Error)",
			method:  http.MethodGet,
			pingErr: errors.New("random error"),
			want:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &PingSrvStub{pingErr: tt.pingErr}
			h := HandlePing(srv, zap.NewNop())

			req := httptest.NewRequest(tt.method, "/ping", nil)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}
