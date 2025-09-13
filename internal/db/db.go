package db

import (
	"database/sql"

	"github.com/alex-storchak/shortener/internal/db/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func GetInstance(cfg *config.Config) (*sql.DB, error) {
	return sql.Open("pgx", cfg.DSN)
}
