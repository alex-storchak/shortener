package model

import "encoding/json"

// URLStorageRecord represents the internal storage structure for URL mappings.
type URLStorageRecord struct {
	OrigURL   string `json:"original_url"` // Original long URL
	ShortID   string `json:"short_url"`    // Generated short identifier
	UserUUID  string `json:"user_uuid"`    // UUID of the user who created the mapping
	IsDeleted bool   `json:"is_deleted"`   // Soft deletion flag
}

// ToJSON serializes the URLStorageRecord to JSON format.
//
// Returns:
//   - []byte: JSON representation of the storage record
//   - error: nil on success, or JSON marshaling error
func (r *URLStorageRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON deserializes JSON data into a URLStorageRecord.
//
// Parameters:
//   - data: JSON byte data to parse
//
// Returns:
//   - error: nil on success, or JSON unmarshaling error
func (r *URLStorageRecord) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}
