package config

import "time"

const (
	DefaultAuthCookieName       = "auth"
	DefaultAuthTokenMaxAge      = 30 * 24 * time.Hour
	DefaultAuthRefreshThreshold = 7 * 24 * time.Hour
	DefaultAuthSecretKey        = "secret"
)

type Config struct {
	AuthCookieName       string        `env:"AUTH_COOKIE_NAME"`
	AuthTokenMaxAge      time.Duration `env:"AUTH_TOKEN_MAX_AGE"`
	AuthRefreshThreshold time.Duration `env:"AUTH_REFRESH_THRESHOLD"`
	SecretKey            string        `env:"AUTH_SECRET_KEY"`
}
