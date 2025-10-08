package middleware

import (
	"errors"
	"net/http"

	"github.com/alex-storchak/shortener/internal/helper"
	mwCfg "github.com/alex-storchak/shortener/internal/middleware/config"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

func AuthMiddleware(
	logger *zap.Logger,
	srv service.AuthMiddlewareService,
	cfg *mwCfg.Config,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string
			if cookie, err := r.Cookie(cfg.AuthCookieName); err == nil {
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

			ctx := helper.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
