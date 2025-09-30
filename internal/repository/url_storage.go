package repository

import (
	"context"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
)

const (
	ShortURLType = `shortURL`
	OrigURLType  = `originalURL`
)

type URLStorage interface {
	Get(url, searchByType string) (string, error)
	Set(origURL, shortURL string) error
	BatchSet(binds *[]model.URLBind) error
	Ping(ctx context.Context) error
	Close() error
}

type DataNotFoundError struct {
	Err error
}

func (e *DataNotFoundError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("no data in the storage for the requested url: %v", e.Err)
	}
	return "no data in the storage for the requested url"
}

func (e *DataNotFoundError) Unwrap() error {
	return e.Err
}

func NewDataNotFoundError(err error) error {
	return &DataNotFoundError{Err: err}
}
