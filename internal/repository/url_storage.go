package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
)

const (
	ShortURLType = `shortURL`
	OrigURLType  = `originalURL`
)

type URLStorage interface {
	Get(ctx context.Context, url, searchByType string) (*model.URLStorageRecord, error)
	Set(ctx context.Context, origURL, shortID, userUUID string) error
	BatchSet(ctx context.Context, records []model.URLStorageRecord) error
	Ping(ctx context.Context) error
	Close() error
	GetByUserUUID(ctx context.Context, userUUID string) ([]*model.URLStorageRecord, error)
	DeleteBatch(ctx context.Context, urls model.URLDeleteBatch) error
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

var (
	ErrDataDeleted = errors.New("data deleted")
)
