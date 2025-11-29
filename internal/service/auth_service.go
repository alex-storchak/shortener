package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/repository"
)

type Claims struct {
	jwt.RegisteredClaims
	UserUUID string
}

type AuthService struct {
	logger *zap.Logger
	us     repository.UserStorage
	secret []byte
	ttl    time.Duration
}

func NewAuthService(logger *zap.Logger, us repository.UserStorage, cfg *config.Auth) *AuthService {
	return &AuthService{
		logger: logger,
		us:     us,
		secret: []byte(cfg.SecretKey),
		ttl:    cfg.TokenMaxAge,
	}
}

func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse auth token: %w", err)
	}

	if !token.Valid {
		return nil, ErrAuthInvalidToken
	}
	a.logger.Debug("auth token is valid")
	return claims, nil
}

func (a *AuthService) ValidateUserUUID(uuid string) (bool, error) {
	if uuid == "" {
		return false, nil
	}
	has, err := a.us.HasByUUID(uuid)
	if err != nil {
		return false, fmt.Errorf("check if user exists: %w", err)
	}
	return has, nil
}

func (a *AuthService) CreateToken(user *model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.ttl)),
		},
		UserUUID: user.UUID,
	})

	tokenString, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return tokenString, nil
}

var (
	ErrAuthInvalidToken = errors.New("invalid auth token")
)
