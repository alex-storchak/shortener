package repository

import (
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type IUserManager interface {
	NewUser() (model.User, error)
}

type UserManager struct {
	logger *zap.Logger
}

func (um *UserManager) NewUser() (model.User, error) {
	return model.User{
		UUID: uuid.NewString(),
	}, nil
}
