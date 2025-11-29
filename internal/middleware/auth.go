package middleware

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

type UserResolver interface {
	ResolveUser(tokenFromCookie string) (user *model.User, newCookie *service.AuthCookieOpts, err error)
}

func NewAuth(logger *zap.Logger, srv UserResolver, cfg config.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string
			if cookie, err := r.Cookie(cfg.CookieName); err == nil {
				token = cookie.Value
			}

			user, newCookie, err := srv.ResolveUser(token)
			if err != nil {
				if errors.Is(err, service.ErrUnauthorized) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				logger.Error("failed to resolve auth user", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if newCookie != nil {
				http.SetCookie(w, &http.Cookie{
					Name:     newCookie.Name,
					Value:    newCookie.Value,
					MaxAge:   newCookie.MaxAge,
					HttpOnly: true,
				})
			}

			ctx := auth.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
