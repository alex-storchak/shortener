package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
)

// UserManager provides user creation and management functionality.
// It generates unique user identifiers and coordinates with storage
// to persist user data.
type UserManager struct {
	logger *zap.Logger
	us     UserStorage
}

// NewUserManager creates a new user manager instance.
//
// Parameters:
//   - logger: structured logger for user management operations
//   - us: user storage for persisting user data
//
// Returns:
//   - *UserManager: configured user manager
func NewUserManager(logger *zap.Logger, us UserStorage) *UserManager {
	return &UserManager{
		logger: logger,
		us:     us,
	}
}

// NewUser creates a new user with a randomly generated UUID and persists it to storage.
// Uses google/uuid package to generate version 4 UUIDs for user identification.
//
// Returns:
//   - *model.User: newly created user object with generated UUID
//   - error: nil on success, or error if UUID generation or storage fails
func (um *UserManager) NewUser() (*model.User, error) {
	user := model.User{
		UUID: uuid.NewString(),
	}
	if err := um.us.Set(context.Background(), &user); err != nil {
		return nil, fmt.Errorf("create new user: %w", err)
	}
	return &user, nil
}
