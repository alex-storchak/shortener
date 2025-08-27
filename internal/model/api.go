package model

type ShortenRequest struct {
	OrigURL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"result"`
}
