package model

//go:generate easyjson -all ./api.go

// ShortenRequest represents the request body for URL shortening operation.
// Used in `POST /api/shorten` endpoint.
type ShortenRequest struct {
	OrigURL string `json:"url"` // Original URL to be shortened
}

// ShortenResponse represents the response body for URL shortening operations.
// Returned by `POST /api/shorten` endpoint with the generated short URL.
type ShortenResponse struct {
	ShortURL string `json:"result"` // Generated short URL
}

// BatchShortenRequestItem represents a single item in batch URL shortening request.
// Contains correlation ID for matching request and response items.
type BatchShortenRequestItem struct {
	CorrelationID string `json:"correlation_id"` // Client-provided identifier for request-response correlation
	OriginalURL   string `json:"original_url"`   // URL to be shortened
}

// BatchShortenRequest represents a collection of URLs for batch shortening.
// Used in `POST /api/shorten/batch` endpoint.
//
//easyjson:json
type BatchShortenRequest []BatchShortenRequestItem

// BatchShortenResponseItem represents a single item in batch URL shortening response.
// Contains the same correlation ID as the request and the generated short URL.
type BatchShortenResponseItem struct {
	CorrelationID string `json:"correlation_id"` // Same identifier from the corresponding request item
	ShortURL      string `json:"short_url"`      // Generated short URL for the original URL
}

// BatchShortenResponse represents the response for batch URL shortening operations.
// Returned by `POST /api/shorten/batch` endpoint.
//
//easyjson:json
type BatchShortenResponse []BatchShortenResponseItem

// UserURLsGetResponseItem represents a single URL record in user URLs response.
// Contains both short and original URLs for user's shortened URLs.
type UserURLsGetResponseItem struct {
	ShortURL string `json:"short_url"`    // Shortened URL identifier
	OrigURL  string `json:"original_url"` // Original full URL
}

// UserURLsGetResponse represents the collection of user's shortened URLs.
// Returned by `GET /api/user/urls` endpoint when user has shortened URLs.
//
//easyjson:json
type UserURLsGetResponse []UserURLsGetResponseItem

// UserURLsDelRequest represents the request body for batch URL deletion.
// Contains a list of short URL identifiers to be marked as deleted.
// Used in `DELETE /api/user/urls` endpoint.
//
//easyjson:json
type UserURLsDelRequest []string

// StatsResponse represents the response for statistics operations.
// Returned by `GET /api/internal/stats` endpoint.
type StatsResponse struct {
	URLsCount  int `json:"urls"`  // Total amount of shortened URLs
	UsersCount int `json:"users"` // Total amount of users
}
