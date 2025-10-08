package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type PgUserDBManager struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewUserDBManager(logger *zap.Logger, db *sql.DB) *PgUserDBManager {
	return &PgUserDBManager{
		logger: logger,
		db:     db,
	}
}

func (m *PgUserDBManager) Close() error {
	return m.db.Close()
}

func (m *PgUserDBManager) HasByUUID(ctx context.Context, uuid string) (bool, error) {
	q := "SELECT id FROM auth_user WHERE user_uuid = $1"
	row := m.db.QueryRowContext(ctx, q, uuid)

	var id int32
	err := row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to scan user query result row: %w", err)
	}
	return true, nil
}

func (m *PgUserDBManager) Persist(ctx context.Context, user *model.User) error {
	q := "INSERT INTO auth_user (user_uuid) VALUES ($1)"
	_, err := m.db.ExecContext(ctx, q, user.UUID)
	if err != nil {
		return fmt.Errorf("failed to persist user (%s) to db: %w", user.UUID, err)
	}
	return nil
}
