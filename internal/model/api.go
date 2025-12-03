package model

//go:generate easyjson -all ./api.go

type ShortenRequest struct {
	OrigURL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"result"`
}

type BatchShortenRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

//easyjson:json
type BatchShortenRequest []BatchShortenRequestItem

type BatchShortenResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//easyjson:json
type BatchShortenResponse []BatchShortenResponseItem

type UserURLsGetResponseItem struct {
	ShortURL string `json:"short_url"`
	OrigURL  string `json:"original_url"`
}

//easyjson:json
type UserURLsGetResponse []UserURLsGetResponseItem

//easyjson:json
type UserURLsDelRequest []string
