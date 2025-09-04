package handler

import (
	"net/http"

	"github.com/alex-storchak/shortener/internal/handler/config"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func Serve(cfg *config.Config, h *Handlers, m *Middlewares) error {
	router := newRouter(h, m)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	return srv.ListenAndServe()
}

func newRouter(h *Handlers, m *Middlewares) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	registerMiddlewares(r, m)
	registerRoutes(r, h)

	return r
}

func registerMiddlewares(router *chi.Mux, middlewares *Middlewares) {
	for _, m := range *middlewares {
		router.Use(m)
	}
}

func registerRoutes(router *chi.Mux, handlers *Handlers) {
	for _, handler := range *handlers {
		for _, route := range handler.Routes() {
			if len(route.Middlewares) > 0 {
				router.With(route.Middlewares...).Method(route.Method, route.Pattern, route.Handler)
			} else {
				router.Method(route.Method, route.Pattern, route.Handler)
			}
		}
	}
}
