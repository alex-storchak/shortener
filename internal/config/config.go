package config

import (
	"time"
)

type Server struct {
	ServerAddr               string        `env:"SERVER_ADDRESS"`
	ShutdownWaitSecsDuration time.Duration `env:"SHUTDOWN_WAIT_SECS_DURATION"`
}

type Handler struct {
	BaseURL string `env:"BASE_URL"`
}

type Logger struct {
	LogLevel string `env:"LOG_LEVEL"`
}

type Repo struct {
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

type DB struct {
	DSN string `env:"DATABASE_DSN"`
}

type Auth struct {
	CookieName       string        `env:"AUTH_COOKIE_NAME"`
	TokenMaxAge      time.Duration `env:"AUTH_TOKEN_MAX_AGE"`
	RefreshThreshold time.Duration `env:"AUTH_REFRESH_THRESHOLD"`
	SecretKey        string        `env:"AUTH_SECRET_KEY"`
}

type Audit struct {
	File             string        `env:"AUDIT_FILE"`
	URL              string        `env:"AUDIT_URL"`
	EventChanSize    int           `env:"AUDIT_EVENT_CHAN_SIZE"`
	HTTPWorkersCount int           `env:"AUDIT_HTTP_WORKERS_COUNT"`
	HTTPTimeout      time.Duration `env:"AUDIT_HTTP_TIMEOUT"`
}

type Config struct {
	Server  Server
	Handler Handler
	Logger  Logger
	Repo    Repo
	DB      DB
	Auth    Auth
	Audit   Audit
}
