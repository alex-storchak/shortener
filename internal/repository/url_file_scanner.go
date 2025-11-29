package repository

import (
	"bufio"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
)

// URLFileScanner provides functionality for scanning and parsing URL records from files.
// It reads files line by line and uses a parser to convert JSON data into URLStorageRecord objects.
type URLFileScanner struct {
	logger *zap.Logger
	parser URLFileRecordParser
}

// NewFileScanner creates a new file scanner instance.
//
// Parameters:
//   - logger: structured logger for logging operations
//   - parser: parser for converting file data to URL records
//
// Returns:
//   - *URLFileScanner: configured file scanner
func NewFileScanner(logger *zap.Logger, parser URLFileRecordParser) *URLFileScanner {
	return &URLFileScanner{
		logger: logger,
		parser: parser,
	}
}

// scan reads the provided file and extracts URL records from each line.
// Skips empty lines and only includes records with non-empty original URLs.
//
// Parameters:
//   - file: file to scan for URL records
//
// Returns:
//   - []model.URLStorageRecord: slice of parsed URL records
//   - error: nil on success, or error if file reading or parsing fails
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
