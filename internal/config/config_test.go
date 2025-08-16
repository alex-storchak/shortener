package config

import (
	"flag"
	"os"
	"testing"

	handlerConfig "github.com/alex-storchak/shortener/internal/handler/config"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		want  Config
	}{
		{
			name:  "get config with default flags",
			flags: []string{},
			want: Config{
				Handler: handlerConfig.Config{
					ServerAddr:       "localhost:8080",
					ShortURLBaseAddr: "http://localhost:8080",
				},
			},
		},
		{
			name: "get config with custom flags",
			flags: []string{
				"-a=example.com:1111",
				"-b=http://example.com:1111",
			},
			want: Config{
				Handler: handlerConfig.Config{
					ServerAddr:       "example.com:1111",
					ShortURLBaseAddr: "http://example.com:1111",
				},
			},
		},
		{
			name: "get config with custom -a (server address) flag",
			flags: []string{
				"-a=example.com:1111",
			},
			want: Config{
				Handler: handlerConfig.Config{
					ServerAddr:       "example.com:1111",
					ShortURLBaseAddr: "http://localhost:8080",
				},
			},
		},
		{
			name: "get config with custom -b (short url base address) flag",
			flags: []string{
				"-b=http://example.com:1111",
			},
			want: Config{
				Handler: handlerConfig.Config{
					ServerAddr:       "localhost:8080",
					ShortURLBaseAddr: "http://example.com:1111",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			oldCommandLine := flag.CommandLine
			defer func() {
				os.Args = oldArgs
				flag.CommandLine = oldCommandLine
			}()

			flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
			testArgs := append([]string{"test"}, tt.flags...)
			os.Args = testArgs

			got := GetConfig()
			assert.Equal(t, tt.want, got)
		})
	}
}
