package model

// User represents a user in the system with a unique identifier.
// The UUID field is used to associate URLs and other resources with specific users.
type User struct {
	UUID string
}
