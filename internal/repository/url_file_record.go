package repository

import (
	"encoding/json"
)

type urlFileRecord struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserUUID    string `json:"user_uuid"`
}

func (r *urlFileRecord) toJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *urlFileRecord) fromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}
