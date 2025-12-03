package repository

import (
	"bufio"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
)

type URLFileScanner struct {
	logger *zap.Logger
	parser URLFileRecordParser
}

func NewFileScanner(logger *zap.Logger, parser URLFileRecordParser) *URLFileScanner {
	return &URLFileScanner{
		logger: logger,
		parser: parser,
	}
}

func (s *URLFileScanner) scan(file *os.File) ([]model.URLStorageRecord, error) {
	scanner := bufio.NewScanner(file)
	var records []model.URLStorageRecord

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		record, err := s.parser.parse(line)
		if err != nil {
			return nil, fmt.Errorf("parse line as record: %w", err)
		}
		if record.OrigURL != "" {
			records = append(records, record)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}
	return records, nil
}
