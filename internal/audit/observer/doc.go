// Package observer provides concrete implementations of the audit.Observer interface.
// It includes observers for different audit event destinations like HTTP servers and local files.
//
// Available Observers:
//   - HTTP: Sends events to remote HTTP servers with configurable timeouts and workers
//   - File: Writes events to local files with proper file locking and append operations
//
// Each observer handles its own connection management, error handling, and resource cleanup.
package observer
