package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

var (
	secretKey           = "secret" // лучше передавать через env
	tokenExpireDuration = time.Hour * 24 * 30

	ErrAuthInvalidToken = errors.New("invalid auth token")
)

type Claims struct {
	jwt.RegisteredClaims
	UserUUID string
}

type AuthService struct {
	logger *zap.Logger
	um     IUserManager
	us     repository.UserStorage
}

func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth token: %w", err)
	}

	if !token.Valid {
		return nil, ErrAuthInvalidToken
	}
	a.logger.Info("auth token is valid")
	return claims, nil
}

func (a *AuthService) ValidateUserUUID(uuid string) (bool, error) {
	if uuid == "" {
		return false, nil
	}
	has, err := a.us.HasByUUID(uuid)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}
	return has, nil
}

func (a *AuthService) CreateToken(user *model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpireDuration)),
		},
		UserUUID: user.UUID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
