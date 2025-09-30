package repository

import (
	"encoding/json"
)

type urlFileRecord struct {
	UUID        uint64 `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserUUID    string `json:"user_uuid,omitempty"`
}

func (r *urlFileRecord) toJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *urlFileRecord) fromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}
