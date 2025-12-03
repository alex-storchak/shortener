package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/model"
)

type UserCreator interface {
	NewUser() (*model.User, error)
}

type AuthCookieOpts struct {
	Name   string
	Value  string
	MaxAge int
}

type AuthUserResolver struct {
	authSrv *AuthService
	uc      UserCreator
	cfg     *config.Auth
}

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

var (
	ErrUnauthorized = errors.New("unauthorized user")
)
