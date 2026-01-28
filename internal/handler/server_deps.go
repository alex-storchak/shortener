package handler

import (
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/interceptor"
	"github.com/alex-storchak/shortener/internal/middleware"
)

// ServerDeps contains dependencies required for HTTP or GRPC server initialization.
type ServerDeps struct {
	Logger              *zap.Logger              // Structured logger for logging operations
	Config              *config.Config           // Application configuration containing server settings and auth configuration
	HTTPUserResolver    middleware.UserResolver  // Service for resolving and validating user authentication in http requests
	GRPCUserResolver    interceptor.UserResolver // Service for resolving and validating user authentication in grpc requests
	ShortenProc         ShortenProcessor         // Processor for plain text URL shortening requests
	ExpandProc          ExpandProcessor          // Processor for expanding short URLs to original URLs
	PingProc            PingProcessor            // Processor for health check requests
	APIShortenProc      APIShortenProcessor      // Processor for JSON API URL shortening requests
	APIShortenBatchProc APIShortenBatchProcessor // Processor for batch URL shortening operations
	APIUserURLsProc     APIUserURLsProcessor     // Processor for user-specific URL management operations
	APIInternalProc     APIInternalProcessor     // Processor for internal stats requests
}
