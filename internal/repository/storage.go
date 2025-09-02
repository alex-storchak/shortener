package repository

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/alex-storchak/shortener/internal/helper"
	"go.uber.org/zap"
)

const (
	ShortURLType = `shortURL`
	OrigURLType  = `originalURL`
)

type URLStorage interface {
	Has(url, urlType string) bool
	Get(url, searchByType string) (string, error)
	Set(originalURL, shortURL string) error
}

type FileURLStorageRecord struct {
	UUID        uint64 `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileURLStorageRecords []FileURLStorageRecord

type FileURLStorage struct {
	dbFile  *os.File
	writer  *bufio.Writer
	records FileURLStorageRecords
	maxUUID uint64
	logger  *zap.Logger
}

func NewFileURLStorage(fileStoragePath string, logger *zap.Logger) (*FileURLStorage, error) {
	logger = logger.With(
		zap.String("component", "storage"),
	)

	fileStoragePath, err := helper.GetAbsFilePath(fileStoragePath)
	logger.Debug("Storage absolute file path", zap.String("path", fileStoragePath))
	if err != nil {
		logger.Error("Can't get absolute file path for file storage path", zap.String("path", fileStoragePath))
		return nil, err
	}

	dbFile, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Error("Can't open file storage path")
		return nil, err
	}
	scanner := bufio.NewScanner(dbFile)

	var records FileURLStorageRecords
	var maxUUID uint64 = 0
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var record FileURLStorageRecord
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, fmt.Errorf("error parsing line '%s': %v", string(line), err)
		}
		maxUUID = record.UUID
		records = append(records, record)
	}
	logger.Debug("Storage maxUUID", zap.Uint64("maxUUID", maxUUID))

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &FileURLStorage{
		dbFile:  dbFile,
		writer:  bufio.NewWriter(dbFile),
		records: records,
		maxUUID: maxUUID,
		logger:  logger,
	}, nil
}

func (s *FileURLStorage) Close() error {
	s.logger.Info("Closing storage file")
	return s.dbFile.Close()
}

func (s *FileURLStorage) nextUUID() uint64 {
	s.maxUUID++
	s.logger.Debug("Next UUID for record", zap.Uint64("UUID", s.maxUUID))
	return s.maxUUID
}

func (s *FileURLStorage) storeRecord(record FileURLStorageRecord) error {
	data, err := json.Marshal(&record)
	if err != nil {
		s.logger.Error("Can't marshal record for store", zap.Error(err))
		return err
	}

	if _, err := s.writer.Write(data); err != nil {
		s.logger.Error("Can't store record", zap.Error(err))
		return err
	}
	if err := s.writer.WriteByte('\n'); err != nil {
		s.logger.Error("Can't add line break in storage", zap.Error(err))
		return err
	}

	s.logger.Info("Stored record",
		zap.Uint64("UUID", record.UUID),
		zap.String("OriginalURL", record.OriginalURL),
		zap.String("ShortURL", record.ShortURL),
	)
	return s.writer.Flush()
}

func (s *FileURLStorage) Has(url, urlType string) bool {
	s.logger.Debug("Checking if url exists",
		zap.String("url", url),
		zap.String("urlType", urlType),
	)
	for _, record := range s.records {
		if urlType == OrigURLType && record.OriginalURL == url {
			return true
		} else if urlType == ShortURLType && record.ShortURL == url {
			return true
		}
	}
	return false
}

func (s *FileURLStorage) Get(url, searchByType string) (string, error) {
	s.logger.Debug("Getting url from storage by type",
		zap.String("url", url),
		zap.String("searchByType", searchByType),
	)
	for _, record := range s.records {
		if searchByType == OrigURLType && record.OriginalURL == url {
			return record.ShortURL, nil
		} else if searchByType == ShortURLType && record.ShortURL == url {
			return record.OriginalURL, nil
		}
	}
	return "", ErrURLStorageDataNotFound
}

func (s *FileURLStorage) Set(originalURL, shortURL string) error {
	s.logger.Debug("Setting url to storage",
		zap.String("originalURL", originalURL),
		zap.String("shortURL", shortURL),
	)
	record := FileURLStorageRecord{
		UUID:        s.nextUUID(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	s.records = append(s.records, record)
	if err := s.storeRecord(record); err != nil {
		return err
	}
	return nil
}

var (
	ErrURLStorageDataNotFound = errors.New("no data in the storage for the requested url")
)
