package config

import "time"

const (
	DefServerAddr               = "localhost:8080"
	DefShutdownWaitSecsDuration = 10 * time.Second

	DefBaseURL = "http://localhost:8080"

	DefLogLevel = "info"

	DefFileStoragePath = "../../../data/file_db.txt"

	DefDatabaseDSN = ""
	MigrationsPath = "file://migrations"

	DefAuthCookieName       = "auth"
	DefAuthTokenMaxAge      = 30 * 24 * time.Hour
	DefAuthRefreshThreshold = 7 * 24 * time.Hour
	DefAuthSecretKey        = "secret"

	DefAuditFile = ""
	DefAuditURL  = ""
)
