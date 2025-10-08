package repository

import (
	"bufio"
	"fmt"
	"os"

	"go.uber.org/zap"
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

func (s *URLFileScanner) scan(file *os.File) (*urlFileRecords, error) {
	scanner := bufio.NewScanner(file)
	var records urlFileRecords
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		record, err := s.parser.parse(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line as record: %w", err)
		}
		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}
	return &records, nil
}
