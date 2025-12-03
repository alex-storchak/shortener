package repository

import (
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
)

// URLFileRecordParser provides functionality for parsing JSON data into URLStorageRecord objects.
// It handles the conversion from byte data to structured records for file storage operations.
type URLFileRecordParser struct{}

// parse converts JSON byte data into a URLStorageRecord.
//
// Parameters:
//   - data: JSON byte data to parse
//
// Returns:
//   - model.URLStorageRecord: parsed URL record
//   - error: nil on success, or error if JSON parsing fails
func (s *URLFileRecordParser) parse(data []byte) (model.URLStorageRecord, error) {
	record := model.URLStorageRecord{}
	if err := record.FromJSON(data); err != nil {
		return model.URLStorageRecord{}, fmt.Errorf("parsing data `%s`: %w", string(data), err)
	}
	return record, nil
}
