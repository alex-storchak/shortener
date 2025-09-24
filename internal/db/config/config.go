package config

const (
	//DefaultDatabaseDSN = "postgres://postgres:StrongPassword01@localhost:54320/shortener?sslmode=disable"
	DefaultDatabaseDSN = "postgres://postgres:StrongPassword01@localhost:54320/postgres?sslmode=disable"
)

type Config struct {
	DSN string `env:"DATABASE_DSN"`
}
