package model

import "encoding/json"

// AuditAction defines the type of action being audited.
type AuditAction string

const (
	// AuditActionShorten represents URL shortening actions.
	AuditActionShorten AuditAction = "shorten"

	// AuditActionFollow represents URL following/redirection actions.
	AuditActionFollow AuditAction = "follow"
)

// AuditEvent represents an audit log entry for tracking system usage.
// Used for monitoring of URL shortening and following activities.
type AuditEvent struct {
	TS      int64       `json:"ts"`                // Unix timestamp of the event
	Action  AuditAction `json:"action"`            // Type of action: shorten or follow
	UserID  string      `json:"user_id,omitempty"` // User identifier, if available
	OrigURL string      `json:"url"`               // Original URL that was processed
}

// ToJSON serializes the AuditEvent to JSON format.
//
// Returns:
//   - []byte: JSON representation of the audit event
//   - error: nil on success, or JSON marshaling error
func (e *AuditEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
