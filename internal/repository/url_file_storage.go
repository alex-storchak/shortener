package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type FileURLStorage struct {
	logger   *zap.Logger
	fileMgr  *FileManager
	fileScnr *URLFileScanner
	records  []*model.URLStorageRecord
	mu       *sync.Mutex
}

func NewFileURLStorage(
	logger *zap.Logger,
	fm *FileManager,
	fs *URLFileScanner,
) (*FileURLStorage, error) {
	storage := &FileURLStorage{
		logger:   logger,
		fileMgr:  fm,
		fileScnr: fs,
		mu:       &sync.Mutex{},
	}

	if err := storage.restoreFromFile(false); err != nil {
		return nil, fmt.Errorf("failed to restore storage from file: %w", err)
	}
	return storage, nil
}

func (s *FileURLStorage) Close() error {
	return s.fileMgr.close()
}

func (s *FileURLStorage) Ping(_ context.Context) error {
	return nil
}

func (s *FileURLStorage) Get(url, searchByType string) (*model.URLStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range s.records {
		if searchByType == OrigURLType && r.OrigURL == url && !r.IsDeleted {
			return r, nil
		} else if searchByType == ShortURLType && r.ShortID == url {
			if r.IsDeleted {
				return nil, ErrDataDeleted
			}
			return r, nil
		}
	}
	return nil, NewDataNotFoundError(nil)
}

func (s *FileURLStorage) Set(r *model.URLStorageRecord) error {
	return s.BatchSet([]*model.URLStorageRecord{r})
}

func (s *FileURLStorage) BatchSet(binds []*model.URLStorageRecord) error {
	if len(binds) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, binds...)
	if err := s.appendToFile(binds); err != nil {
		// rollback
		s.records = s.records[:len(s.records)-len(binds)]
		return fmt.Errorf("failed to persist records batch to file: %w", err)
	}
	return nil
}

func (s *FileURLStorage) GetByUserUUID(userUUID string) ([]*model.URLStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var records []*model.URLStorageRecord
	for _, r := range s.records {
		if r.UserUUID == userUUID && !r.IsDeleted {
			records = append(records, r)
		}
	}
	return records, nil
}

func (s *FileURLStorage) DeleteBatch(urls model.URLDeleteBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ProcessMemDeleteBatch(s.records, urls)
	if err := s.saveToFile(); err != nil {
		return fmt.Errorf("failed to save records to file: %w", err)
	}
	return nil
}

func (s *FileURLStorage) appendToFile(records []*model.URLStorageRecord) error {
	_, err := s.fileMgr.openForAppend(false)
	if err != nil {
		return fmt.Errorf("failed to open file for append: %w", err)
	}
	defer s.fileMgr.close()

	if err := s.writeRecords(records); err != nil {
		return fmt.Errorf("failed to write records to file: %w", err)
	}

	return nil
}

func (s *FileURLStorage) saveToFile() error {
	_, err := s.fileMgr.openForWrite(false)
	if err != nil {
		return fmt.Errorf("failed to open file for write: %w", err)
	}
	defer s.fileMgr.close()

	if err = s.writeRecords(s.records); err != nil {
		return fmt.Errorf("failed to write records to file: %w", err)
	}
	return nil
}

func (s *FileURLStorage) writeRecords(records []*model.URLStorageRecord) error {
	for _, r := range records {
		data, err := r.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to convert record to json for store: %w", err)
		}

		if err := s.fileMgr.writeData(data); err != nil {
			return fmt.Errorf("mgr failed to persist record to file: %w", err)
		}
	}
	return nil
}

func (s *FileURLStorage) restoreFromFile(useDefault bool) error {
	file, err := s.fileMgr.openForAppend(useDefault)
	if err != nil && !useDefault {
		s.logger.Warn("failed to restore from requested file, trying default: ", zap.Error(err))
		if err := s.restoreFromFile(true); err != nil {
			return fmt.Errorf("failed to restore from default file: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to open default file: %w", err)
	}
	defer s.fileMgr.close()

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
