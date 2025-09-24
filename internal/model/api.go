package model

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

type BatchShortenResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
