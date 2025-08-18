package config

import (
	"testing"

	handlerConfig "github.com/alex-storchak/shortener/internal/handler/config"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name string
		want Config
	}{
		{
			name: "get default config",
			want: Config{
				Handler: handlerConfig.Config{
					ServerHost: "localhost",
					ServerPort: "8080",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetConfig()
			assert.Equal(t, got, tt.want)
		})
	}
}
