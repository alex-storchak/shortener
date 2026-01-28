package config

import "time"

// Default configuration values used when no environment variable or flag is provided.
//
// Server defaults
const (
	// DefServerAddr - Default server address
	DefServerAddr = "localhost:8080"
	// DefGRPCServerAddr - Default gRPC server address
	DefGRPCServerAddr = ":50051"
	// DefEnableHTTPS - Default flag value for using HTTPS in server
	DefEnableHTTPS = false
	// DefSSLCertPath - Default path to SSL certificate file
	DefSSLCertPath = ""
	// DefSSLKeyPath - Default path to SSL key file
	DefSSLKeyPath = ""
	// DefShutdownWaitSecsDuration - Default graceful shutdown timeout
	DefShutdownWaitSecsDuration = 10 * time.Second
)

// Handler defaults
const (
	// DefBaseURL - Default base URL for short links
	DefBaseURL = "http://localhost:8080"
)

// Logger defaults
const (
	// DefLogLevel - Default log level
	DefLogLevel = "info"
)

// Repository defaults
const (
	// DefFileStoragePath - Default file storage path
	DefFileStoragePath = "../../../data/file_db.txt"
)

// Database defaults
const (
	// DefDatabaseDSN - Default database DSN (empty = no database)
	DefDatabaseDSN = ""
	// DefMigrationsPath - Path to database migration files
	DefMigrationsPath = "file://migrations"
)

// Authentication defaults
const (
	// DefAuthCookieName - Default authentication cookie name
	DefAuthCookieName = "authorization"
	// DefAuthTokenMaxAge - Default token max age (30 days)
	DefAuthTokenMaxAge = 30 * 24 * time.Hour
	// DefAuthRefreshThreshold -  Default refresh threshold (7 days)
	DefAuthRefreshThreshold = 7 * 24 * time.Hour
	// DefAuthSecretKey - Default JWT secret key
	DefAuthSecretKey = "secret"
)

// Audit system defaults
const (
	// DefAuditFile - Default audit file path (empty = disabled)
	DefAuditFile = ""
	// DefAuditURL - Default audit server URL (empty = disabled)
	DefAuditURL = ""
	// DefAuditEventChanSize - Default audit event channel size
	DefAuditEventChanSize = 1000
	// DefAuditHTTPWorkersCount - Default number of audit HTTP workers
	DefAuditHTTPWorkersCount = 8
	// DefAuditHTTPTimeout - Default audit HTTP request timeout
	DefAuditHTTPTimeout = 3 * time.Second
)
