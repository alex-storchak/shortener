// Package config provides configuration management for the URL shortener service.
// It handles loading configuration from multiple sources with proper precedence:
// environment variables override command-line flags, which override default values.
//
// Configuration Sources (in order of precedence):
//   - Environment variables (highest priority)
//   - Command-line flags
//   - Default values (lowest priority)
//
// The package supports configuration for:
//   - Server settings (address, shutdown timeout)
//   - Handler settings (base URL for short links)
//   - Logging configuration
//   - Storage options (file path, database DSN)
//   - Authentication (JWT, cookies)
//   - Audit system (file logging, remote server)
//
// Usage:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal("Failed to load configuration:", err)
//	}
package config
