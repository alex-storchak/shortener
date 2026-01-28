package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadFromFile(path string, cfg *Config) error {
	if path == "" {
		return nil
	}
	jc, err := readJSONConfigFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	applyJSONConfig(cfg, jc)
	return nil
}

func readJSONConfigFile(path string) (*JSONConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}
	var jc JSONConfig
	if err := json.Unmarshal(b, &jc); err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", path, err)
	}
	return &jc, nil
}

func applyJSONConfig(cfg *Config, jc *JSONConfig) {
	if jc == nil {
		return
	}

	// Server
	if jc.ServerAddress != nil {
		cfg.Server.ServerAddr = *jc.ServerAddress
	}
	if jc.EnableHTTPS != nil {
		cfg.Server.EnableHTTPS = *jc.EnableHTTPS
	}
	if jc.SSLCertPath != nil {
		cfg.Server.SSLCertPath = *jc.SSLCertPath
	}
	if jc.SSLKeyPath != nil {
		cfg.Server.SSLKeyPath = *jc.SSLKeyPath
	}
	if jc.ShutdownWaitSecsDuration != nil {
		cfg.Server.ShutdownWaitSecsDuration = *jc.ShutdownWaitSecsDuration
	}
	if jc.TrustedSubnet != nil {
		cfg.Server.TrustedSubnet = *jc.TrustedSubnet
	}

	// Handler
	if jc.BaseURL != nil {
		cfg.Handler.BaseURL = *jc.BaseURL
	}

	// Logger
	if jc.LogLevel != nil {
		cfg.Logger.LogLevel = *jc.LogLevel
	}

	// Repo
	if jc.FileStoragePath != nil {
		cfg.Repo.FileStoragePath = *jc.FileStoragePath
	}

	// DB
	if jc.DatabaseDSN != nil {
		cfg.DB.DSN = *jc.DatabaseDSN
	}
	if jc.DatabaseMigrationsPath != nil {
		cfg.DB.MigrationsPath = *jc.DatabaseMigrationsPath
	}

	// Auth
	if jc.AuthCookieName != nil {
		cfg.Auth.CookieName = *jc.AuthCookieName
	}
	if jc.AuthTokenMaxAge != nil {
		cfg.Auth.TokenMaxAge = *jc.AuthTokenMaxAge
	}
	if jc.AuthRefreshThreshold != nil {
		cfg.Auth.RefreshThreshold = *jc.AuthRefreshThreshold
	}
	if jc.AuthSecretKey != nil {
		cfg.Auth.SecretKey = *jc.AuthSecretKey
	}

	// Audit
	if jc.AuditFile != nil {
		cfg.Audit.File = *jc.AuditFile
	}
	if jc.AuditURL != nil {
		cfg.Audit.URL = *jc.AuditURL
	}
	if jc.AuditEventChanSize != nil {
		cfg.Audit.EventChanSize = *jc.AuditEventChanSize
	}
	if jc.AuditHTTPWorkersCount != nil {
		cfg.Audit.HTTPWorkersCount = *jc.AuditHTTPWorkersCount
	}
	if jc.AuditHTTPTimeout != nil {
		cfg.Audit.HTTPTimeout = *jc.AuditHTTPTimeout
	}
}
