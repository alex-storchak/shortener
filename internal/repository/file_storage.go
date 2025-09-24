package repository

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type fileRecords []fileRecord

type FileURLStorage struct {
	logger   *zap.Logger
	fileMgr  *FileManager
	fileScnr *FileScanner
	uuidMgr  *UUIDManager
	records  *fileRecords
}

func NewFileURLStorage(
	logger *zap.Logger,
	fm *FileManager,
	fs *FileScanner,
	um *UUIDManager,
) (*FileURLStorage, error) {
	storage := &FileURLStorage{
		logger:   logger,
		fileMgr:  fm,
		fileScnr: fs,
		uuidMgr:  um,
	}

	if err := storage.restoreFromFile(false); err != nil {
		return nil, fmt.Errorf("failed to restore storage from file: %w", err)
	}
	storage.initUUIDMgr()
	return storage, nil
}

func (s *FileURLStorage) Close() error {
	return s.fileMgr.close()
}

func (s *FileURLStorage) Ping(_ context.Context) error {
	return nil
}

func (s *FileURLStorage) persistToFile(record fileRecord) error {
	data, err := record.toJSON()
	if err != nil {
		return fmt.Errorf("failed to convert record to json for store: %w", err)
	}
	if err := s.fileMgr.writeData(data); err != nil {
		return fmt.Errorf("mgr failed to persist record to file: %w", err)
	}
	s.logger.Info("Stored record",
		zap.Uint64("UUID", record.UUID),
		zap.String("OriginalURL", record.OriginalURL),
		zap.String("ShortURL", record.ShortURL),
	)
	return nil
}

func (s *FileURLStorage) Get(url, searchByType string) (string, error) {
	for _, record := range *s.records {
		if searchByType == OrigURLType && record.OriginalURL == url {
			return record.ShortURL, nil
		} else if searchByType == ShortURLType && record.ShortURL == url {
			return record.OriginalURL, nil
		}
	}
	return "", NewDataNotFoundError(nil)
}

func (s *FileURLStorage) Set(origURL, shortURL string) error {
	record := fileRecord{
		UUID:        s.uuidMgr.next(),
		ShortURL:    shortURL,
		OriginalURL: origURL,
	}
	*s.records = append(*s.records, record)
	if err := s.persistToFile(record); err != nil {
		return fmt.Errorf("failed to persist record `%v` to file: %w", record, err)
	}
	return nil
}

func (s *FileURLStorage) BatchSet(binds *[]URLBind) error {
	for _, b := range *binds {
		if err := s.Set(b.OrigURL, b.ShortID); err != nil {
			return fmt.Errorf("failed to set record in storage: %w", err)
		}
	}
	return nil
}

func (s *FileURLStorage) initUUIDMgr() {
	var maxUUID uint64 = 0
	for _, rec := range *s.records {
		if rec.UUID > maxUUID {
			maxUUID = rec.UUID
		}
	}
	s.uuidMgr.init(maxUUID)
}

func (s *FileURLStorage) restoreFromFile(useDefault bool) error {
	file, err := s.fileMgr.open(useDefault)
	if err != nil && !useDefault {
		s.logger.Warn("failed to restore from requested file, trying default: ", zap.Error(err))
		if err := s.restoreFromFile(true); err != nil {
			return fmt.Errorf("failed to restore from default file: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to open default file: %w", err)
	}

	records, err := s.fileScnr.scan(file)
	if err != nil && !useDefault {
		s.logger.Warn("Can't scan data from requested file, trying default: ", zap.Error(err))
		if err := s.fileMgr.close(); err != nil {
			return fmt.Errorf("failed to close requested file: %w", err)
		}
		if err := s.restoreFromFile(true); err != nil {
			if err := s.fileMgr.close(); err != nil {
				return fmt.Errorf("failed to close requested file: %w", err)
			}
			return fmt.Errorf("failed to restore from default file: %w", err)
		}
		return nil
	} else if err != nil {
		if cErr := s.fileMgr.close(); cErr != nil {
			return fmt.Errorf("failed to close default file: %w", cErr)
		}
		return fmt.Errorf("failed to scan data from default file: %w", err)
	}

	s.records = records
	return nil
}
