// Package audit provides an auditing system for tracking URL shortening operations.
// It implements the Observer pattern to distribute audit events to multiple outputs simultaneously.
//
// Architecture:
//   - EventManager: Central dispatcher that receives events and distributes to observers
//   - Observer: Interface for different audit event destinations (HTTP, file, etc.)
//   - HTTP Observer: Sends events to remote audit servers via HTTP
//   - File Observer: Writes events to local files in JSON format
//
// Features:
//   - Asynchronous event processing with buffered channels
//   - Multiple output destinations with configurable limits
//   - Graceful shutdown with proper resource cleanup
//   - Configurable queue sizes and timeouts
//   - Thread-safe operations
//
// Usage:
//
//	observers, err := audit.InitObservers(cfg.Audit, logger)
//	em := audit.NewEventManager(observers, cfg.Audit, logger)
//	defer em.Close(context.Background())
//
//	em.Publish(model.AuditEvent{
//	    TS:     time.Now().Unix(),
//	    Action: model.AuditActionShorten,
//	    UserID: "user123",
//	    OrigURL: "https://example.com",
//	})
package audit
