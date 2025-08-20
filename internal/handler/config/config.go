package config

const (
	DefaultServerAddr = "localhost:8080"
	DefaultBaseURL    = "http://localhost:8080"
)

type Config struct {
	ServerAddr string `env:"SERVER_ADDRESS"`
	BaseURL    string `env:"BASE_URL"`
}
