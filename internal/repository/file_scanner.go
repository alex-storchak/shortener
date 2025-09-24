package repository

import (
	"bufio"
	"fmt"
	"os"

	"go.uber.org/zap"
)

type FileScanner struct {
	logger *zap.Logger
	parser FileRecordParser
}

func NewFileScanner(logger *zap.Logger, parser FileRecordParser) *FileScanner {
	return &FileScanner{
		logger: logger,
		parser: parser,
	}
}

func (s *FileScanner) scan(file *os.File) (*fileRecords, error) {
	scanner := bufio.NewScanner(file)
	var records fileRecords
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
