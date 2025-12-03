// Package middleware provides HTTP middleware components for the shortener service.
// It includes authentication, compression, logging, and other cross-cutting concerns
// that can be composed around HTTP handlers.
//
// The middleware follows the standard http.Handler pattern and can be easily
// integrated with any compatible HTTP router or framework.
//
// Key components:
//   - NewAuth: Authentication middleware that resolves users from JWT tokens
//   - NewGzip: Compression middleware that handles gzip encoding
//   - NewRequestLogger: Request logging middleware with detailed metrics
//
// # Usage Example
//
//	router := chi.NewRouter()
//	router.Use(middleware.NewRequestLogger(logger))
//	router.Use(middleware.NewGzip(logger))
//	router.Use(middleware.NewAuth(logger, userResolver, authConfig))
//	router.Mount("/", handler)
package middleware
