package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
)

var ErrEmptySSLCertPath = errors.New("empty SSL certificate path")
var ErrEmptySSLKeyPath = errors.New("empty SSL key path")

// NewRouter creates and configures the HTTP router with all application routes and middleware.
// It sets up the complete routing structure including authentication, logging, compression,
// and all URL shortening endpoints.
//
// Parameters:
//   - h: ServerDeps containing all dependencies required for HTTP server initialization
//
// Returns:
//   - http.Handler: Configured HTTP router with all middleware and routes
func NewRouter(h *ServerDeps) http.Handler {
	mux := chi.NewRouter()
	addRoutes(mux, h)
	return mux
}

// Serve initializes and starts the HTTP server.
// It creates an HTTP server with the provided configuration and router,
// starts listening for incoming connections in a goroutine,
// and verifies successful server startup.
//
// Parameters:
//   - cfg: Server configuration containing address, SSL settings, and other server options
//   - logger: Structured logger for logging operations
//   - router: HTTP handler with all configured routes and middleware
//
// Returns:
//   - *http.Server: Started HTTP server instance for later shutdown management
//   - error: nil on successful startup, or error if server fails to start
//
// Note:
//   - The server runs in a separate goroutine and continues running after this function returns
//   - Caller must manage server lifecycle including graceful shutdown
//   - SSL certificate and key paths are validated when HTTPS is enabled
func Serve(cfg config.Server, logger *zap.Logger, router http.Handler) (*http.Server, error) {
	httpServer := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting server", zap.String("server address", cfg.ServerAddr))
		var err error
		if cfg.EnableHTTPS {
			err = startHTTPS(cfg, httpServer)
		} else {
			err = startHTTP(httpServer)
		}
		if err != nil {
			errCh <- err
		}
	}()

	// Проверяем, успешно ли запустился сервер
	select {
	case err := <-errCh:
		return nil, fmt.Errorf("start server: %w", err)
	case <-time.After(time.Second):
		return httpServer, nil
	}
}

func startHTTP(s *http.Server) error {
	err := s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("starting server: %w", err)
	}
	return nil
}

func startHTTPS(cfg config.Server, s *http.Server) error {
	if cfg.SSLCertPath == "" {
		return ErrEmptySSLCertPath
	}
	if cfg.SSLKeyPath == "" {
		return ErrEmptySSLKeyPath
	}
	err := s.ListenAndServeTLS(cfg.SSLCertPath, cfg.SSLKeyPath)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("starting tls server: %w", err)
	}
	return nil
}
