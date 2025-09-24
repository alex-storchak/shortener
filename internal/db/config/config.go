package config

const (
	DefaultDatabaseDSN = ""
	MigrationsPath     = "file://migrations"
)

type Config struct {
	DSN string `env:"DATABASE_DSN"`
}
