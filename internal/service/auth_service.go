package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/repository"
)

// Claims represents the JWT claims structure used for authentication.
// It extends jwt.RegisteredClaims with a UserUUID field to identify the user.
type Claims struct {
	jwt.RegisteredClaims
	UserUUID string
}

// AuthService provides JWT token creation and validation functionality.
// It handles token signing, verification, and user validation.
type AuthService struct {
	logger *zap.Logger
	us     repository.UserStorage
	secret []byte
	ttl    time.Duration
}

// NewAuthService creates a new AuthService instance with the specified dependencies.
//
// Parameters:
//   - logger: structured logger for logging operations
//   - us: user storage for validating user existence
//   - cfg: authentication configuration including secret key and token TTL
//
// Returns:
//   - *AuthService: configured authentication service
func NewAuthService(logger *zap.Logger, us repository.UserStorage, cfg *config.Auth) *AuthService {
	return &AuthService{
		logger: logger,
		us:     us,
		secret: []byte(cfg.SecretKey),
		ttl:    cfg.TokenMaxAge,
	}
}

// ValidateToken parses and validates a JWT token string.
// It verifies the token signature, expiration, and returns the claims if valid.
//
// Parameters:
//   - tokenString: JWT token string to validate
//
// Returns:
//   - *Claims: decoded claims from the token if valid
//   - error: nil if token is valid, or validation error
//
// Errors:
//   - ErrAuthInvalidToken: when token signature is invalid or token is expired
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

// ValidateUserUUID checks if a user with the given UUID exists in the system.
//
// Parameters:
//   - uuid: user UUID to validate
//
// Returns:
//   - bool: true if user exists, false otherwise
//   - error: nil on success, or storage error if validation fails
func (a *AuthService) ValidateUserUUID(uuid string) (bool, error) {
	if uuid == "" {
		return false, nil
	}
	has, err := a.us.HasByUUID(context.Background(), uuid)
	if err != nil {
		return false, fmt.Errorf("check if user exists: %w", err)
	}
	return has, nil
}

// CreateToken generates a new JWT token for the specified user.
// The token includes the user's UUID and has an expiration time based on the service TTL.
//
// Parameters:
//   - user: user object for which to create the token
//
// Returns:
//   - string: signed JWT token string
//   - error: nil on success, or token signing error
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

// Common authentication errors
var (
	// ErrAuthInvalidToken is returned when a JWT token is invalid, expired, or has an invalid signature.
	ErrAuthInvalidToken = errors.New("invalid auth token")
)
