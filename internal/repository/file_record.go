package repository

import (
	"encoding/json"
)

type fileRecord struct {
	UUID        uint64 `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (r *fileRecord) toJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *fileRecord) fromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}
