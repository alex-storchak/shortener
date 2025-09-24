package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
