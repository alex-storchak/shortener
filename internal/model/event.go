package model

import "encoding/json"

type AuditAction string

const (
	AuditActionShorten AuditAction = "shorten"
	AuditActionFollow  AuditAction = "follow"
)

type AuditEvent struct {
	TS      int64       `json:"ts"`
	Action  AuditAction `json:"action"`
	UserID  string      `json:"user_id,omitempty"`
	OrigURL string      `json:"url"`
}

func (e *AuditEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
