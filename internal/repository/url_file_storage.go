package repository

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
)

// URLFileManager defines the interface for file management operations used by file storage.
type URLFileManager interface {
	OpenForAppend(useDefault bool) (*os.File, error)
	OpenForWrite(useDefault bool) (*os.File, error)
	Close() error
	WriteData(data []byte) error
}

// FileURLStorage provides a file-based implementation of URLStorage.
// It persists URL data to disk in JSON format and restores it on initialization.
// The storage maintains an in-memory index of records for fast lookups
// and synchronizes changes to disk.
//
// This implementation provides crash recovery by restoring from the storage file
// and supports fallback to a default file if the primary file is unavailable.
type FileURLStorage struct {
	logger    *zap.Logger
	fileMgr   URLFileManager
	fileScnr  *URLFileScanner
	records   []model.URLStorageRecord
	mu        *sync.Mutex
	closeOnce sync.Once
}

// NewFileURLStorage creates a new file-based URL storage instance.
// It automatically restores existing URL data from the storage file on initialization.
//
// Parameters:
//   - logger: structured logger for logging operations
//   - fm: file manager for file operations
//   - fs: file scanner for reading URL records from file
//
// Returns:
//   - *FileURLStorage: configured file-based URL storage
//   - error: nil on success, or error if file restoration fails
func NewFileURLStorage(
	logger *zap.Logger,
	fm URLFileManager,
	fs *URLFileScanner,
) (*FileURLStorage, error) {
	storage := &FileURLStorage{
		logger:   logger,
		fileMgr:  fm,
		fileScnr: fs,
		mu:       &sync.Mutex{},
	}

	if err := storage.restoreFromFile(false); err != nil {
		return nil, fmt.Errorf("restore storage from file: %w", err)
	}
	return storage, nil
}

// Close releases file resources used by the storage.
//
// Returns:
//   - error: nil on success, or error if file closure fails
func (s *FileURLStorage) Close() error {
	return s.fileMgr.Close()
}

// Ping always returns nil for file storage as file operations are local.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//
// Returns:
//   - error: always returns nil
func (s *FileURLStorage) Ping(_ context.Context) error {
	return nil
}

// Get retrieves a URL record from file storage based on search type.
// Uses the in-memory index for fast lookups after initial file restoration.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//   - url: URL to search for
//   - searchByType: type of search (ShortURLType or OrigURLType)
//
// Returns:
//   - *model.URLStorageRecord: found record or nil if not found
//   - error: nil on success, or ErrDataDeleted if URL is deleted
func (s *FileURLStorage) Get(_ context.Context, url, searchByType string) (*model.URLStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range s.records {
		if searchByType == OrigURLType && r.OrigURL == url && !r.IsDeleted {
			return &r, nil
		} else if searchByType == ShortURLType && r.ShortID == url {
			if r.IsDeleted {
				return nil, ErrDataDeleted
			}
			return &r, nil
		}
	}
	return nil, NewDataNotFoundError(nil)
}

// Set stores a single URL mapping in file storage.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - OrigURL: original URL to be shortened
//   - ShortID: generated short identifier
//   - UserUUID: UUID of the user who created the mapping
//
// Returns:
//   - error: nil on success, or error if file write fails
func (s *FileURLStorage) Set(ctx context.Context, OrigURL, ShortID, UserUUID string) error {
	return s.BatchSet(ctx, []model.URLStorageRecord{
		{
			OrigURL:  OrigURL,
			ShortID:  ShortID,
			UserUUID: UserUUID,
		},
	})
}

// BatchSet stores multiple URL mappings in file storage and persists them to disk.
// Implements atomic-like behavior with rollback on write failure.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//   - binds: slice of URL storage records to persist
//
// Returns:
//   - error: nil on success, or error if file write fails
func (s *FileURLStorage) BatchSet(_ context.Context, binds []model.URLStorageRecord) error {
	if len(binds) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, binds...)
	if err := s.appendToFile(binds); err != nil {
		// rollback
		s.records = s.records[:len(s.records)-len(binds)]
		return fmt.Errorf("persist records batch to file: %w", err)
	}
	return nil
}

// GetByUserUUID retrieves all non-deleted URL mappings for a specific user.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//   - userUUID: UUID of the user to retrieve URLs for
//
// Returns:
//   - []*model.URLStorageRecord: slice of URL records belonging to the user
//   - error: nil on success
func (s *FileURLStorage) GetByUserUUID(_ context.Context, userUUID string) ([]*model.URLStorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := make([]*model.URLStorageRecord, 0, 100)
	for _, r := range s.records {
		if r.UserUUID == userUUID && !r.IsDeleted {
			records = append(records, &r)
		}
	}
	return records, nil
}

// DeleteBatch marks multiple URLs as deleted and persists the changes to disk.
//
// Parameters:
//   - ctx: context for cancellation and timeouts (not used)
//   - urls: batch of URLs to delete with user identifiers
//
// Returns:
//   - error: nil on success, or error if file write fails
func (s *FileURLStorage) DeleteBatch(_ context.Context, urls model.URLDeleteBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ProcessMemDeleteBatch(s.records, urls)
	if err := s.saveToFile(); err != nil {
		return fmt.Errorf("save records to file: %w", err)
	}
	return nil
}

// appendToFile appends new records to the storage file.
func (s *FileURLStorage) appendToFile(records []model.URLStorageRecord) error {
	_, err := s.fileMgr.OpenForAppend(false)
	if err != nil {
		return fmt.Errorf("open file for append: %w", err)
	}
	defer s.fileMgr.Close()

	if err := s.writeRecords(records); err != nil {
		return fmt.Errorf("write records to file: %w", err)
	}

	return nil
}

// saveToFile writes all records to the storage file, overwriting existing content.
func (s *FileURLStorage) saveToFile() error {
	_, err := s.fileMgr.OpenForWrite(false)
	if err != nil {
		return fmt.Errorf("open file for write: %w", err)
	}
	defer s.fileMgr.Close()

	if err = s.writeRecords(s.records); err != nil {
		return fmt.Errorf("write records to file: %w", err)
	}
	return nil
}

// writeRecords converts records to JSON and writes them to the file.
func (s *FileURLStorage) writeRecords(records []model.URLStorageRecord) error {
	for _, r := range records {
		data, err := r.ToJSON()
		if err != nil {
			return fmt.Errorf("convert record to json for store: %w", err)
		}

		if err := s.fileMgr.WriteData(data); err != nil {
			return fmt.Errorf("mgr persist record to file: %w", err)
		}
	}
	return nil
}

// restoreFromFile reads the storage file and rebuilds the in-memory record index.
// Supports fallback to default file if primary file is unavailable.
func (s *FileURLStorage) restoreFromFile(useDefault bool) error {
	file, err := s.fileMgr.OpenForAppend(useDefault)
	if err != nil && !useDefault {
		s.logger.Warn("failed to restore from requested file, trying default: ", zap.Error(err))
		if err := s.restoreFromFile(true); err != nil {
			return fmt.Errorf("restore from default file: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("open default file: %w", err)
	}
	defer s.fileMgr.Close()

	records, err := s.fileScnr.scan(file)
	if err != nil && !useDefault {
		s.logger.Warn("Can't scan data from requested file, trying default: ", zap.Error(err))
		if err := s.fileMgr.Close(); err != nil {
			return fmt.Errorf("close requested file: %w", err)
		}
		if err := s.restoreFromFile(true); err != nil {
			if err := s.fileMgr.Close(); err != nil {
				return fmt.Errorf("close requested file: %w", err)
			}
			return fmt.Errorf("restore from default file: %w", err)
		}
		return nil
	} else if err != nil {
		if cErr := s.fileMgr.Close(); cErr != nil {
			return fmt.Errorf("close default file: %w", cErr)
		}
		return fmt.Errorf("scan data from default file: %w", err)
	}

	s.records = records
	return nil
}
