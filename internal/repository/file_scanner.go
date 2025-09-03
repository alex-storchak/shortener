package repository

import (
	"bufio"
	"os"

	"go.uber.org/zap"
)

type FileScanner struct {
	logger *zap.Logger
	parser FileRecordParser
}

func NewFileScanner(logger *zap.Logger, parser FileRecordParser) *FileScanner {
	logger = logger.With(
		zap.String("component", "file scanner"),
	)
	return &FileScanner{
		logger: logger,
		parser: parser,
	}
}

func (s *FileScanner) scan(file *os.File) (fileRecords, error) {
	scanner := bufio.NewScanner(file)
	var records fileRecords
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		record, err := s.parser.parse(line)
		if err != nil {
			s.logger.Error("failed to parse record",
				zap.Error(err),
				zap.String("line", string(line)),
			)
			return nil, err
		}
		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		s.logger.Error("scanner error", zap.Error(err))
		return nil, err
	}
	return records, nil
}
