package service

import "strings"

// URLBuilder provides functionality for constructing full URLs based on a base URL and short identifier.
// The structure ensures proper URL part concatenation with correct slash handling.
type URLBuilder struct {
	baseURL string
}

// NewURLBuilder creates a new URLBuilder instance with the specified base URL.
// The base URL is automatically normalized - trailing slashes are removed and a single slash is appended.
// This ensures consistent full URL formation when using the Build method.
//
// Example:
//
//	builder := NewURLBuilder("https://example.com")
//	builder := NewURLBuilder("https://example.com/")
//	// Both will create a builder with baseURL = "https://example.com/"
func NewURLBuilder(base string) *URLBuilder {
	base = strings.TrimRight(base, "/")
	withSlash := base + "/"
	return &URLBuilder{baseURL: withSlash}
}

// Build constructs a full URL by combining the base URL and short identifier.
// The method automatically handles cases where shortID starts with a slash, preventing duplicate slashes.
//
// Parameters:
//   - shortID: short identifier to append to the base URL. Can be an empty string.
//
// Returns:
//   - string: complete URL formed as baseURL + shortID
//
// Examples:
//
//	builder := NewURLBuilder("https://example.com/")
//	url1 := builder.Build("abc")       // "https://example.com/abc"
//	url2 := builder.Build("/abc")      // "https://example.com/abc"
//	url3 := builder.Build("")          // "https://example.com/"
func (u *URLBuilder) Build(shortID string) string {
	if shortID != "" && shortID[0] == '/' {
		// capacity: baseURL + shortID without the leading '/'
		b := make([]byte, 0, len(u.baseURL)+len(shortID)-1)
		b = append(b, u.baseURL...)
		b = append(b, shortID[1:]...)
		return string(b)
	}

	b := make([]byte, 0, len(u.baseURL)+len(shortID))
	b = append(b, u.baseURL...)
	b = append(b, shortID...)
	return string(b)
}
