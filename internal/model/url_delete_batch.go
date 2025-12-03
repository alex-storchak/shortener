package model

// URLToDelete represents a single URL deletion request with user authorization.
type URLToDelete struct {
	UserUUID string // UUID of the user requesting deletion
	ShortID  string // Short URL identifier to delete
}

// URLDeleteBatch represents a collection of URLs to be deleted in batch.
type URLDeleteBatch []URLToDelete
