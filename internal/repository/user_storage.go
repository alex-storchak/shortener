package repository

import "github.com/alex-storchak/shortener/internal/model"

type UserStorage interface {
	HasByUUID(uuid string) (bool, error)
	Set(user *model.User) error
	Close() error
}
