package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/model"
)

// UserCreator defines the interface for creating new users in the system.
// Implementations should generate unique user identifiers and persist them.
type UserCreator interface {
	NewUser() (*model.User, error)
}

// AuthCookieOpts represents the configuration for authentication cookies.
// It is used to set HTTP cookies with JWT tokens for client authentication.
type AuthCookieOpts struct {
	Name   string // Cookie name (e.g., "auth")
	Value  string // JWT token value
	MaxAge int    // Cookie expiration in seconds
}

// AuthUserResolver handles user authentication resolution and token management.
// It coordinates between AuthService for token operations and UserCreator for user management.
// The resolver provides automatic token refresh and new user creation when needed.
type AuthUserResolver struct {
	authSrv *AuthService
	uc      UserCreator
	cfg     *config.Auth
}

// NewAuthUserResolver creates a new AuthUserResolver instance.
//
// Parameters:
//   - a: authentication service for token validation and creation
//   - uc: user creator for generating new users when needed
//   - cfg: authentication configuration including cookie settings and refresh thresholds
//
// Returns:
//   - *AuthUserResolver: configured authentication resolver
func NewAuthUserResolver(
	a *AuthService,
	uc UserCreator,
	cfg *config.Auth,
) *AuthUserResolver {
	return &AuthUserResolver{
		authSrv: a,
		uc:      uc,
		cfg:     cfg,
	}
}

// ResolveUserGRPC resolves a user based on the provided authentication token.
// If token is empty or invalid, returns ErrUnauthorized.
//
// Parameters:
//   - token: JWT token string from client request (can be empty)
//
// Returns:
//   - *model.User: resolved user object (never nil on success)
//   - error: authentication error if resolution fails
func (s *AuthUserResolver) ResolveUserGRPC(token string) (*model.User, error) {
	if token == "" {
		return nil, ErrUnauthorized
	}

	claims, err := s.authSrv.ValidateToken(token)
	if err != nil {
		return nil, ErrUnauthorized
	}

	ok, uErr := s.authSrv.ValidateUserUUID(claims.UserUUID)
	if uErr != nil {
		return nil, fmt.Errorf("validate user uuid: %w", uErr)
	}
	if !ok {
		return nil, ErrUnauthorized
	}

	user := &model.User{UUID: claims.UserUUID}

	return user, nil
}

// ResolveUser resolves a user based on the provided authentication token.
// It handles three main scenarios:
//   - No token provided: creates new user and issues token
//   - Invalid token: creates new user and issues token
//   - Valid token: returns existing user, refreshing token if near expiration
//
// Parameters:
//   - token: JWT token string from client request (can be empty)
//
// Returns:
//   - *model.User: resolved user object (never nil on success)
//   - *AuthCookieOpts: cookie options if a new cookie should be set (new user or token refresh)
//   - error: authentication error if resolution fails
//
// The method ensures that a valid user is always returned when no error occurs.
// Cookie options are returned when:
//   - A new user is created
//   - An existing token is refreshed due to nearing expiration
func (s *AuthUserResolver) ResolveUser(token string) (*model.User, *AuthCookieOpts, error) {
	if token == "" {
		return s.issueNewUserAndCookie()
	}

	claims, err := s.authSrv.ValidateToken(token)
	if err != nil {
		return s.issueNewUserAndCookie()
	}

	ok, uErr := s.authSrv.ValidateUserUUID(claims.UserUUID)
	if uErr != nil {
		return nil, nil, fmt.Errorf("validate user uuid: %w", uErr)
	}
	if !ok {
		return nil, nil, ErrUnauthorized
	}

	user := &model.User{UUID: claims.UserUUID}

	if claims.ExpiresAt.Before(time.Now().Add(s.cfg.RefreshThreshold)) {
		newToken, err := s.authSrv.CreateToken(user)
		if err != nil {
			return nil, nil, err
		}
		return user, &AuthCookieOpts{
			Name:   s.cfg.CookieName,
			Value:  newToken,
			MaxAge: int(s.cfg.TokenMaxAge.Seconds()),
		}, nil
	}

	return user, nil, nil
}

// issueNewUserAndCookie handles the creation of a new user and associated authentication token.
// This is used when no valid authentication token is available.
//
// Returns:
//   - *model.User: newly created user
//   - *AuthCookieOpts: cookie configuration with new authentication token
//   - error: nil on success, or error if user creation or token generation fails
func (s *AuthUserResolver) issueNewUserAndCookie() (*model.User, *AuthCookieOpts, error) {
	user, err := s.uc.NewUser()
	if err != nil {
		return nil, nil, fmt.Errorf("create new user: %w", err)
	}
	token, err := s.authSrv.CreateToken(user)
	if err != nil {
		return nil, nil, fmt.Errorf("create token for new user: %w", err)
	}
	return user, &AuthCookieOpts{
		Name:   s.cfg.CookieName,
		Value:  token,
		MaxAge: int(s.cfg.TokenMaxAge.Seconds()),
	}, nil
}

// Common authentication errors
var (
	// ErrUnauthorized is returned when a user UUID in a token is not found in the system.
	ErrUnauthorized = errors.New("unauthorized user")
)
