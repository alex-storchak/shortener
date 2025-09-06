package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_validateMethod(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		allowed string
		wantErr bool
	}{
		{
			name:    "allowed method matches",
			method:  http.MethodPost,
			allowed: http.MethodPost,
			wantErr: false,
		},
		{
			name:    "method does not match allowed",
			method:  http.MethodGet,
			allowed: http.MethodPost,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMethod(tt.method, tt.allowed)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Nil(t, err)
		})
	}
}

func Test_validateContentType(t *testing.T) {
	tests := []struct {
		name    string
		ct      string
		allowed string
		wantErr bool
	}{
		{
			name:    "allowed content-type matches",
			ct:      "application/json",
			allowed: "application/json",
			wantErr: false,
		},
		{
			name:    "content-type does not match allowed",
			ct:      "text/plain",
			allowed: "application/json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContentType(tt.ct, tt.allowed)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Nil(t, err)
		})
	}
}
