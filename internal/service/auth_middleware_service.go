package service

import (
	"errors"
	"fmt"
	"time"

	mwcfg "github.com/alex-storchak/shortener/internal/middleware/config"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/repository"
)

type AuthCookieOpts struct {
	Name   string
	Value  string
	MaxAge int
}

type AuthMiddlewareService interface {
	ResolveUser(tokenFromCookie string) (user *model.User, newCookie *AuthCookieOpts, err error)
}

type authMiddlewareService struct {
	authSrv *AuthService
	uc      repository.UserCreator
	cfg     *mwcfg.Config
}

func NewAuthMiddlewareService(
	authSrv *AuthService,
	uc repository.UserCreator,
	cfg *mwcfg.Config,
) AuthMiddlewareService {
	return &authMiddlewareService{
		authSrv: authSrv,
		uc:      uc,
		cfg:     cfg,
	}
}

func (s *authMiddlewareService) ResolveUser(token string) (*model.User, *AuthCookieOpts, error) {
	if token == "" {
		return s.issueNewUserAndCookie()
	}

	claims, err := s.authSrv.ValidateToken(token)
	if err != nil {
		return s.issueNewUserAndCookie()
	}

	ok, uErr := s.authSrv.ValidateUserUUID(claims.UserUUID)
	if uErr != nil {
		return nil, nil, fmt.Errorf("failed to validate user uuid: %w", uErr)
	}
	if !ok {
		return nil, nil, ErrUnauthorized
	}

	user := &model.User{UUID: claims.UserUUID}

	if claims.ExpiresAt.Before(time.Now().Add(s.cfg.AuthRefreshThreshold)) {
		newToken, err := s.authSrv.CreateToken(user)
		if err != nil {
			return nil, nil, err
		}
		return user, &AuthCookieOpts{
			Name:   s.cfg.AuthCookieName,
			Value:  newToken,
			MaxAge: int(s.cfg.AuthTokenMaxAge.Seconds()),
		}, nil
	}

	return user, nil, nil
}

func (s *authMiddlewareService) issueNewUserAndCookie() (*model.User, *AuthCookieOpts, error) {
	user, err := s.uc.NewUser()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new user: %w", err)
	}
	token, err := s.authSrv.CreateToken(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create token for new user: %w", err)
	}
	return user, &AuthCookieOpts{
		Name:   s.cfg.AuthCookieName,
		Value:  token,
		MaxAge: int(s.cfg.AuthTokenMaxAge.Seconds()),
	}, nil
}

var (
	ErrUnauthorized = errors.New("unauthorized user")
)
