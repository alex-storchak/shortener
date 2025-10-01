package repository

import (
	"fmt"
)

type URLFileRecordParser struct{}

func (s *URLFileRecordParser) parse(data []byte) (urlFileRecord, error) {
	record := urlFileRecord{}
	if err := record.fromJSON(data); err != nil {
		return record, fmt.Errorf("error parsing data `%s`: %w", string(data), err)
	}
	return record, nil
}
