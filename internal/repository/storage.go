package repository

import "context"

const (
	ShortURLType = `shortURL`
	OrigURLType  = `originalURL`
)

type URLBind struct {
	OrigURL string
	ShortID string
}

type URLStorage interface {
	Get(url, searchByType string) (string, error)
	Set(origURL, shortURL string) error
	BatchSet(binds *[]URLBind) error
	Ping(ctx context.Context) error
	Close() error
}
