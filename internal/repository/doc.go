// Package repository provides data storage abstractions and implementations for the URL shortener.
// It defines interfaces and concrete implementations for persisting URL mappings and user data.
//
// # Storage Interfaces
//
// Core abstractions for data persistence:
//   - URLStorage: interface for URL mapping operations (CRUD, batch, user-specific)
//   - UserStorage: interface for user data management
//
// # Storage Implementations
//
// Multiple storage backends with consistent interfaces:
//   - MemoryURLStorage: in-memory storage for testing/development
//   - FileURLStorage: file-based persistence with JSON serialization
//   - DBURLStorage: PostgreSQL-based storage with transaction support
//   - MemoryUserStorage/FileUserStorage/DBUserStorage: corresponding user storage implementations
//
// # Common Patterns
//
// The package implements several important patterns:
//   - DataNotFoundError: standardized error handling for missing data
//   - Soft deletion: URL records are marked deleted rather than removed
//   - Batch operations: efficient processing of multiple items
//
// # File Storage Support
//
// Additional components for file-based storage:
//   - URLFileScanner: reads and parses URL records from files
//   - URLFileRecordParser: converts JSON data to storage records
//
// # Error Handling
//
// Consistent error types across implementations:
//   - DataNotFoundError: when requested data doesn't exist
//   - ErrDataDeleted: when accessing soft-deleted URLs
//
// Package repository provides the data access layer with pluggable storage backends,
// allowing the application to use memory, file, or database storage based on configuration.
package repository
