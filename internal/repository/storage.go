package repository

const (
	ShortURLType = `shortURL`
	OrigURLType  = `originalURL`
)

type URLStorage interface {
	Get(url, searchByType string) (string, error)
	Set(origURL, shortURL string) error
}
