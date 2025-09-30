package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alex-storchak/shortener/internal/helper"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

func createUser(um repository.IUserManager) (*model.User, error) {
	user, err := um.NewUser()
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}
	return user, nil
}

func AuthMiddleware(logger *zap.Logger, authSrv *service.AuthService, um repository.IUserManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user *model.User

			cookie, err := r.Cookie("auth")
			if err == nil { // cookie exists
				token := cookie.Value
				claims, vErr := authSrv.ValidateToken(token)
				if vErr != nil {
					logger.Info("failed to validate token, create new user", zap.Error(vErr))
				} else {
					ok, uErr := authSrv.ValidateUserUUID(claims.UserUUID)
					if uErr != nil {
						logger.Error("failed to validate user uuid", zap.Error(err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					if !ok {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}

					user = &model.User{UUID: claims.UserUUID}
					if claims.ExpiresAt.Before(time.Now().Add(time.Hour * 24 * 7)) { // if tokens expires in 7 days
						logger.Info("increasing token's expire time")
					} else {
						http.SetCookie(w, &http.Cookie{
							Name:     "auth",
							Value:    token,
							HttpOnly: true,
							MaxAge:   30 * 24 * 60 * 60, // 30 дней
						})
						ctx := context.WithValue(r.Context(), helper.UserCtxKey{}, user)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// cookie doesn't exist or need to extend expire time
			if user == nil {
				var uErr error
				user, uErr = createUser(um)
				if uErr != nil {
					logger.Error("failed to create new user", zap.Error(uErr))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			token, tErr := authSrv.CreateToken(user)
			if tErr != nil {
				logger.Error("failed to create auth token", zap.Error(tErr))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "auth",
				Value:    token,
				HttpOnly: true,
				MaxAge:   30 * 24 * 60 * 60, // 30 дней
			})
			ctx := context.WithValue(r.Context(), helper.UserCtxKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
