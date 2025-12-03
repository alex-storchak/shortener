// Package model provides the data structures and types used throughout the application.
// It defines the core domain models, API request/response structures, and storage representations.
//
// # Core Domain Models
//
// The package includes domain models:
//   - User: represents system users with unique identifiers
//   - URLStorageRecord: internal storage structure for URL mappings
//   - URLToDelete and URLDeleteBatch: for batch deletion operations
//
// # API Models
//
// Structured types for REST API communication:
//   - ShortenRequest/ShortenResponse: for single URL shortening
//   - BatchShortenRequest/BatchShortenResponse: for batch URL operations
//   - UserURLsGetResponse: for retrieving user's shortened URLs
//   - UserURLsDelRequest: for batch URL deletion requests
//
// # Audit System
//
// Support for request auditing and monitoring:
//   - AuditEvent: captures action details for analytics
//   - AuditAction: defines possible audit actions (shorten, follow)
//
// # JSON Support
//
// All models support JSON serialization/deserialization with optional easyjson optimizations.
// The package includes code generation directives for high-performance JSON handling.
//
// Package model serves as the central data contract between different layers of the application,
// ensuring consistent data representation across HTTP handlers, business logic, and storage layers.
package model
