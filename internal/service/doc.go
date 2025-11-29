// Package service provides the core business logic for the URL shortener service.
// It includes functionality for URL shortening, user authentication, and service orchestration.
//
// # Main Components
//
// URL Shortening:
//   - Shortener: Main service for URL shortening operations
//   - URLBuilder: Constructs full URLs from short identifiers
//   - IDGenerator: Interface for generating unique short IDs
//
// Authentication:
//   - AuthService: JWT token creation and validation
//   - AuthUserResolver: User authentication and token management
//   - UserManager: User creation and management
//
// # Key Features
//
//   - Shorten individual URLs and batches of URLs
//   - Extract original URLs from short identifiers
//   - User authentication with JWT tokens
//   - Automatic token refresh
//   - User-specific URL management
//   - Batch URL deletion with soft delete
//   - Health checking and readiness probes
//
// # Interfaces
//
// The package defines several interfaces to support different implementations:
//   - URLShortener: Core URL shortening operations
//   - PingableURLShortener: URL shortener with health checking
//   - IDGenerator: Short ID generation
//   - Pinger: Service readiness checking
//   - UserCreator: User creation
//
// # Error Handling
//
// The package defines common errors for consistent error handling:
//   - ErrURLAlreadyExists: When a URL already has a short identifier
//   - ErrEmptyInputURL: When an empty URL is provided
//   - ErrEmptyInputBatch: When an empty batch is provided
//   - ErrUnauthorized: When user authentication fails
//   - ErrAuthInvalidToken: When JWT token validation fails
//
// # Dependencies
//
// The service layer depends on:
//   - repository package for data persistence
//   - model package for data structures
//   - config package for configuration
//   - zap for structured logging
//   - jwt for token handling
package service
