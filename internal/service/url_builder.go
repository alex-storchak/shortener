package service

import "strings"

type URLBuilder struct {
	baseURL string
}

func NewURLBuilder(base string) *URLBuilder {
	base = strings.TrimRight(base, "/")
	withSlash := base + "/"
	return &URLBuilder{baseURL: withSlash}
}

func (u *URLBuilder) Build(shortID string) string {
	if shortID != "" && shortID[0] == '/' {
		// capacity: baseURL + shortID без начального '/'
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
