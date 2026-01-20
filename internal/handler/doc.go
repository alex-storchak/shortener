// Package handler provides HTTP handlers and routing for the URL shortener service.
// It implements REST API endpoints for URL shortening, expansion, batch operations,
// user management, and health checks.
//
// Key Features:
//   - Middleware for authentication, logging, compression, and recovery
//   - Support for both JSON API and plain text endpoints
//   - Audit event publishing for security and monitoring
//   - Graceful server shutdown and lifecycle management
//
// Endpoints:
//   - POST /                   - Shorten URL (text/plain)
//   - GET  /{id}               - Expand short URL to original
//   - GET  /ping               - Health check
//   - POST /api/shorten        - Shorten URL (JSON API)
//   - POST /api/shorten/batch  - Batch URL shortening
//   - GET  /api/user/urls      - Get user's URLs
//   - DELETE /api/user/urls    - Delete user's URLs
//   - GET  /api/internal/stats - Get amount of URLs and users in storage
//
// Middleware:
//   - Request logging with structured logging
//   - Gzip compression for requests and responses
//   - JWT-based authentication
//   - Panic recovery
//   - Content type validation
//
// The package uses the Chi router for flexible routing and middleware composition.
package handler
