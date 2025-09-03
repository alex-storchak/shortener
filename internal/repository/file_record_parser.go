package repository

import (
	"fmt"
)

type FileRecordParser struct{}

func (s *FileRecordParser) parse(data []byte) (fileRecord, error) {
	record := fileRecord{}
	if err := record.fromJSON(data); err != nil {
		return record, fmt.Errorf("error parsing data '%s': %v", string(data), err)
	}
	return record, nil
}
