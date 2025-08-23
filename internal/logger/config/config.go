package config

const (
	DefaultLogLevel = "info"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL"`
}
