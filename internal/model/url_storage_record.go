package model

import "encoding/json"

type URLStorageRecord struct {
	OrigURL   string `json:"original_url"`
	ShortID   string `json:"short_url"`
	UserUUID  string `json:"user_uuid"`
	IsDeleted bool   `json:"is_deleted"`
}

func (r *URLStorageRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *URLStorageRecord) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}
