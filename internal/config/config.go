package config

import (
	"time"
)

// Server contains configuration for HTTP server settings.
type Server struct {
	ServerAddr               string        `env:"SERVER_ADDRESS"`              // Address to bind the HTTP server (e.g., "localhost:8080")
	EnableHTTPS              bool          `env:"ENABLE_HTTPS"`                // HTTPS server mode flag
	SSLCertPath              string        `env:"SSL_CERT_PATH"`               // Path to SSL certificate
	SSLKeyPath               string        `env:"SSL_KEY_PATH"`                // Path to SSL key
	ShutdownWaitSecsDuration time.Duration `env:"SHUTDOWN_WAIT_SECS_DURATION"` // Graceful shutdown timeout duration
	TrustedSubnet            string        `env:"TRUSTED_SUBNET"`              // Trusted subnet for internal stats requests (CIDR notation, e.g., "127.0.0.1/32")
}

// Reset set all fields of Server to default values
func (s *Server) Reset() {
	s.ServerAddr = DefServerAddr
	s.EnableHTTPS = DefEnableHTTPS
	s.SSLCertPath = DefSSLCertPath
	s.SSLKeyPath = DefSSLKeyPath
	s.ShutdownWaitSecsDuration = DefShutdownWaitSecsDuration
}

// Handler contains configuration for HTTP handler settings.
type Handler struct {
	BaseURL string `env:"BASE_URL"` // Base URL for generated short URLs (e.g., "http://localhost:8080")
}

// Reset set all fields of Handler to default values
func (h *Handler) Reset() {
	h.BaseURL = DefBaseURL
}

// Logger contains configuration for logging settings.
type Logger struct {
	LogLevel string `env:"LOG_LEVEL"` // Log level (debug, info, warn, error)
}

// Reset set all fields of Logger to default values
func (l *Logger) Reset() {
	l.LogLevel = DefLogLevel
}

// Repo contains configuration for repository/storage settings.
type Repo struct {
	FileStoragePath string `env:"FILE_STORAGE_PATH"` // Path to file storage for URL data
}

// Reset set all fields of Repo to default values
func (r *Repo) Reset() {
	r.FileStoragePath = DefFileStoragePath
}

// DB contains configuration for database settings.
type DB struct {
	DSN            string `env:"DATABASE_DSN"`             // PostgreSQL connection string
	MigrationsPath string `env:"DATABASE_MIGRATIONS_PATH"` // Path to database migrations
}

// Reset set all fields of DB to default values
func (d *DB) Reset() {
	d.DSN = DefDatabaseDSN
	d.MigrationsPath = DefMigrationsPath
}

// Auth contains configuration for authentication settings.
type Auth struct {
	CookieName       string        `env:"AUTH_COOKIE_NAME"`       // Name of the authentication cookie
	TokenMaxAge      time.Duration `env:"AUTH_TOKEN_MAX_AGE"`     // Maximum age of authentication tokens
	RefreshThreshold time.Duration `env:"AUTH_REFRESH_THRESHOLD"` // Threshold for token refresh
	SecretKey        string        `env:"AUTH_SECRET_KEY"`        // Secret key for JWT token signing
}

// Reset set all fields of Auth to default values
func (a *Auth) Reset() {
	a.CookieName = DefAuthCookieName
	a.TokenMaxAge = DefAuthTokenMaxAge
	a.RefreshThreshold = DefAuthRefreshThreshold
	a.SecretKey = DefAuthSecretKey
}

// Audit contains configuration for audit system settings.
type Audit struct {
	File             string        `env:"AUDIT_FILE"`               // File path for audit logging
	URL              string        `env:"AUDIT_URL"`                // Remote server URL for audit events
	EventChanSize    int           `env:"AUDIT_EVENT_CHAN_SIZE"`    // Size of audit event channel buffer
	HTTPWorkersCount int           `env:"AUDIT_HTTP_WORKERS_COUNT"` // Number of HTTP workers for audit events
	HTTPTimeout      time.Duration `env:"AUDIT_HTTP_TIMEOUT"`       // HTTP timeout for audit requests
}

// Reset set all fields of Audit to default values
func (a *Audit) Reset() {
	a.File = DefAuditFile
	a.URL = DefAuditURL
	a.EventChanSize = DefAuditEventChanSize
	a.HTTPWorkersCount = DefAuditHTTPWorkersCount
	a.HTTPTimeout = DefAuditHTTPTimeout
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

// Reset set all fields of Config to default values.
func (c *Config) Reset() {
	c.Server.Reset()
	c.Handler.Reset()
	c.Logger.Reset()
	c.Repo.Reset()
	c.DB.Reset()
	c.Auth.Reset()
	c.Audit.Reset()
}

// JSONConfig is a plain structure of config from JSON file
type JSONConfig struct {
	// Server
	ServerAddress            *string        `json:"server_address"`
	EnableHTTPS              *bool          `json:"enable_https"`
	SSLCertPath              *string        `json:"ssl_cert_path"`
	SSLKeyPath               *string        `json:"ssl_key_path"`
	ShutdownWaitSecsDuration *time.Duration `json:"shutdown_wait_secs_duration"`
	TrustedSubnet            *string        `json:"trusted_subnet"`

	// Handler
	BaseURL *string `json:"base_url"`

	// Logger
	LogLevel *string `json:"log_level"`

	// Repo
	FileStoragePath *string `json:"file_storage_path"`

	// DB
	DatabaseDSN            *string `json:"database_dsn"`
	DatabaseMigrationsPath *string `json:"database_migrations_path"`

	// Auth
	AuthCookieName       *string        `json:"auth_cookie_name"`
	AuthTokenMaxAge      *time.Duration `json:"auth_token_max_age"`
	AuthRefreshThreshold *time.Duration `json:"auth_refresh_threshold"`
	AuthSecretKey        *string        `json:"auth_secret_key"`

	// Audit
	AuditFile             *string        `json:"audit_file"`
	AuditURL              *string        `json:"audit_url"`
	AuditEventChanSize    *int           `json:"audit_event_chan_size"`
	AuditHTTPWorkersCount *int           `json:"audit_http_workers_count"`
	AuditHTTPTimeout      *time.Duration `json:"audit_http_timeout"`
}
