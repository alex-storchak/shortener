package config

const (
	//DefaultDatabaseDSN = "postgres://shortener:StrongPassword02@localhost:54320/shortener?sslmode=disable&search_path=public"
	DefaultDatabaseDSN = ""
)

type Config struct {
	DSN string `env:"DATABASE_DSN"`
}
