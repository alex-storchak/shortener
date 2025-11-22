package repository

import (
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserManager struct {
	logger *zap.Logger
	us     UserStorage
}

func NewUserManager(logger *zap.Logger, us UserStorage) *UserManager {
	return &UserManager{
		logger: logger,
		us:     us,
	}
}

func (um *UserManager) NewUser() (*model.User, error) {
	user := model.User{
		UUID: uuid.NewString(),
	}
	if err := um.us.Set(&user); err != nil {
		return nil, fmt.Errorf("create new user: %w", err)
	}
	return &user, nil
}
