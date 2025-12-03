// Package factory provides storage factory implementations for creating configured storage instances.
// It implements the Factory pattern to abstract storage creation and supports multiple backends.
//
// # Storage Factory Interface
//
// The core StorageFactory interface:
//   - MakeURLStorage(): creates URL storage instances
//   - MakeUserStorage(): creates user storage instances
//
// # Factory Implementations
//
// Concrete factories for different storage types:
//   - MemoryStorageFactory: creates in-memory storage instances
//   - FileStorageFactory: creates file-based storage with automatic data restoration
//   - DBStorageFactory: creates database storage with connection management
//
// # Automatic Factory Selection
//
// The package provides intelligent factory selection based on configuration:
//   - Database storage: when DSN is configured (highest priority)
//   - File storage: when file storage path is configured
//   - Memory storage: as fallback when no persistence is configured
//
// # Configuration-Based Initialization
//
// Factories are initialized with application configuration:
//   - Database factories establish connections and run migrations
//   - File factories set up file managers and scanners
//   - Memory factories require minimal configuration
//
// # Usage
//
// Typical usage involves creating a factory through NewStorageFactory() and then
// using it to create both URL and user storage instances that work together.
//
// Package factory simplifies storage configuration and initialization, providing
// a consistent way to create properly configured storage instances throughout the application.
package factory
