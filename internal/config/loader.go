package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Load initializes and loads the application configuration from all available sources.
// It parses command-line flags first, then environment variables, with environment
// variables taking precedence over flag values.
//
// Parameters:
//   - args - Slice with os.Args
//   - lookupEnv - Function for checking whether env variables has been set (ex. os.LookupEnv)
//
// Returns:
//   - *Config: Fully populated configuration structure
//   - error: nil on success, or error if environment parsing fails
func Load(
	args []string,
	lookupEnv func(string) (string, bool),
) (*Config, error) {
	cfg := &Config{}
	cfg.Reset()

	cfgFilePath, err := getCfgFilePath(args, lookupEnv)
	if err != nil {
		return cfg, fmt.Errorf("get config file path: %w", err)
	}
	if err := loadFromFile(cfgFilePath, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	parseFlags(cfg)
	err = parseEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("parse env vars: %w", err)
	}
	return cfg, nil
}

// parseFlags parses command-line flags and sets configuration values.
// Flags provide the base configuration that can be overridden by environment variables.
//
// Parameters:
//   - cfg: Pointer to Config structure to populate with flag values
func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.Server.ServerAddr, "a", cfg.Server.ServerAddr, "address of HTTP server")
	flag.StringVar(&cfg.Server.GRPCServerAddr, "grpc-server-addr", cfg.Server.GRPCServerAddr, "address of gRPC server")
	flag.BoolVar(&cfg.Server.EnableHTTPS, "s", cfg.Server.EnableHTTPS, "enable HTTPS in HTTP server")
	flag.StringVar(&cfg.Server.SSLCertPath, "ssl-cert", cfg.Server.SSLCertPath, "path to SSL certificate file")
	flag.StringVar(&cfg.Server.SSLKeyPath, "ssl-key", cfg.Server.SSLKeyPath, "path to SSL key file")
	flag.DurationVar(&cfg.Server.ShutdownWaitSecsDuration, "shutdown-wait-secs-duration", cfg.Server.ShutdownWaitSecsDuration, "server shutdown wait seconds duration")
	flag.StringVar(&cfg.Server.TrustedSubnet, "t", cfg.Server.TrustedSubnet, "trusted subnet for internal stats requests (CIDR notation, e.g., \"127.0.0.1/32\")")

	flag.StringVar(&cfg.Handler.BaseURL, "b", cfg.Handler.BaseURL, "base URL of short url service")

	flag.StringVar(&cfg.Logger.LogLevel, "l", cfg.Logger.LogLevel, "log level")

	flag.StringVar(&cfg.Repo.FileStoragePath, "f", cfg.Repo.FileStoragePath, "db storage file path")

	flag.StringVar(&cfg.DB.DSN, "d", cfg.DB.DSN, "postgres database DSN")
	flag.StringVar(&cfg.DB.MigrationsPath, "m", cfg.DB.MigrationsPath, "postgres database migrations path")

	flag.StringVar(&cfg.Auth.CookieName, "auth-cookie-name", cfg.Auth.CookieName, "auth cookie name")
	flag.DurationVar(&cfg.Auth.TokenMaxAge, "auth-token-max-age", cfg.Auth.TokenMaxAge, "auth token max age in hours")
	flag.DurationVar(&cfg.Auth.RefreshThreshold, "auth-refresh-threshold", cfg.Auth.RefreshThreshold, "auth refresh threshold in hours")
	flag.StringVar(&cfg.Auth.SecretKey, "auth-secret-key", cfg.Auth.SecretKey, "auth JWT secret key")

	flag.StringVar(&cfg.Audit.File, "audit-file", cfg.Audit.File, "audit log file path")
	flag.StringVar(&cfg.Audit.URL, "audit-url", cfg.Audit.URL, "full URL of audit server")
	flag.IntVar(&cfg.Audit.EventChanSize, "audit-event-chan-size", cfg.Audit.EventChanSize, "audit event chan size")
	flag.IntVar(&cfg.Audit.HTTPWorkersCount, "audit-http-workers-count", cfg.Audit.HTTPWorkersCount, "audit http workers count")
	flag.DurationVar(&cfg.Audit.HTTPTimeout, "audit-http-timeout", cfg.Audit.HTTPTimeout, "audit http timeout")

	flag.Parse()
}

// parseEnv parses environment variables and overrides any previously set flag values.
// Environment variables have the highest precedence in the configuration hierarchy.
//
// Parameters:
//   - cfg: Pointer to Config structure to populate with environment values
//
// Returns:
//   - error: nil on success, or error if environment parsing fails
func parseEnv(cfg *Config) error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("parse env with lib: %w", err)
	}
	return nil
}

func getCfgFilePath(args []string, lookup func(string) (string, bool)) (string, error) {
	if fromEnv, ok := lookup("CONFIG"); ok {
		return fromEnv, nil
	}

	fromFlag, _, err := findConfigFlag(args)
	if err != nil {
		return "", fmt.Errorf("find config flag: %w", err)
	}
	return fromFlag, nil
}

// findConfigFlag extracts config path from args.
// Supports: -c value, -c=value
//
// Returns:
//   - path: Path to config file
//   - found: Whether flag presents in args
//   - error: Error if parsing fails
func findConfigFlag(args []string) (string, bool, error) {
	for i := 0; i < len(args); i++ {
		a := args[i]

		if a == "-c" {
			if i+1 >= len(args) {
				return "", true, fmt.Errorf("%s requires a value", a)
			}
			return args[i+1], true, nil
		}

		if strings.HasPrefix(a, "-c=") {
			return strings.TrimPrefix(a, "-c="), true, nil
		}
	}
	return "", false, nil
}
