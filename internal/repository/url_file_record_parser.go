package repository

import (
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
)

type URLFileRecordParser struct{}

func (s *URLFileRecordParser) parse(data []byte) (*model.URLStorageRecord, error) {
	record := model.URLStorageRecord{}
	if err := record.FromJSON(data); err != nil {
		return nil, fmt.Errorf("parsing data `%s`: %w", string(data), err)
	}
	return &record, nil
}
