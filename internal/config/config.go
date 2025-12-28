package config

import (
	"time"
)

// Server contains configuration for HTTP server settings.
type Server struct {
	// Address to bind the HTTP server (e.g., "localhost:8080")
	ServerAddr string `env:"SERVER_ADDRESS"`
	// HTTPS server mode flag
	EnableHTTPS bool `env:"ENABLE_HTTPS"`
	// Path to SSL certificate
	SSLCertPath string `env:"SSL_CERT_PATH"`
	// Path to SSL key
	SSLKeyPath string `env:"SSL_KEY_PATH"`
	// Graceful shutdown timeout duration
	ShutdownWaitSecsDuration time.Duration `env:"SHUTDOWN_WAIT_SECS_DURATION"`
}

// Handler contains configuration for HTTP handler settings.
type Handler struct {
	// Base URL for generated short URLs (e.g., "http://localhost:8080")
	BaseURL string `env:"BASE_URL"`
}

// Logger contains configuration for logging settings.
type Logger struct {
	LogLevel string `env:"LOG_LEVEL"` // Log level (debug, info, warn, error)
}

// Repo contains configuration for repository/storage settings.
type Repo struct {
	FileStoragePath string `env:"FILE_STORAGE_PATH"` // Path to file storage for URL data
}

// DB contains configuration for database settings.
type DB struct {
	DSN            string `env:"DATABASE_DSN"`             // PostgreSQL connection string
	MigrationsPath string `env:"DATABASE_MIGRATIONS_PATH"` // Path to database migrations
}

// Auth contains configuration for authentication settings.
type Auth struct {
	CookieName       string        `env:"AUTH_COOKIE_NAME"`       // Name of the authentication cookie
	TokenMaxAge      time.Duration `env:"AUTH_TOKEN_MAX_AGE"`     // Maximum age of authentication tokens
	RefreshThreshold time.Duration `env:"AUTH_REFRESH_THRESHOLD"` // Threshold for token refresh
	SecretKey        string        `env:"AUTH_SECRET_KEY"`        // Secret key for JWT token signing
}

// Audit contains configuration for audit system settings.
type Audit struct {
	File             string        `env:"AUDIT_FILE"`               // File path for audit logging
	URL              string        `env:"AUDIT_URL"`                // Remote server URL for audit events
	EventChanSize    int           `env:"AUDIT_EVENT_CHAN_SIZE"`    // Size of audit event channel buffer
	HTTPWorkersCount int           `env:"AUDIT_HTTP_WORKERS_COUNT"` // Number of HTTP workers for audit events
	HTTPTimeout      time.Duration `env:"AUDIT_HTTP_TIMEOUT"`       // HTTP timeout for audit requests
}

// Config represents the complete application configuration.
// It aggregates all configuration sections into a single structure.
type Config struct {
	Server  Server  // HTTP server configuration
	Handler Handler // HTTP handler configuration
	Logger  Logger  // Logging configuration
	Repo    Repo    // Repository/storage configuration
	DB      DB      // Database configuration
	Auth    Auth    // Authentication configuration
	Audit   Audit   // Audit system configuration
}
