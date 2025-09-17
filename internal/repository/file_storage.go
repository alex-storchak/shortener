package repository

import (
	"errors"

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
	logger = logger.With(
		zap.String("component", "storage"),
	)

	storage := &FileURLStorage{
		logger:   logger,
		fileMgr:  fm,
		fileScnr: fs,
		uuidMgr:  um,
	}

	if err := storage.restoreFromFile(false); err != nil {
		logger.Error("failed to restore storage from file", zap.Error(err))
		return nil, err
	}
	storage.initUUIDMgr()
	return storage, nil
}

func (s *FileURLStorage) Close() error {
	s.logger.Info("Closing storage file")
	return s.fileMgr.close()
}

func (s *FileURLStorage) persistToFile(record fileRecord) error {
	data, err := record.toJSON()
	if err != nil {
		s.logger.Error("Can't prepare record for store", zap.Error(err))
		return err
	}

	if err := s.fileMgr.writeData(data); err != nil {
		s.logger.Error("Can't persist record to file", zap.Error(err))
		return err
	}

	s.logger.Info("Stored record",
		zap.Uint64("UUID", record.UUID),
		zap.String("OriginalURL", record.OriginalURL),
		zap.String("ShortURL", record.ShortURL),
	)

	return nil
}

func (s *FileURLStorage) Get(url, searchByType string) (string, error) {
	s.logger.Debug("Getting url from storage by type",
		zap.String("url", url),
		zap.String("searchByType", searchByType),
	)
	for _, record := range *s.records {
		if searchByType == OrigURLType && record.OriginalURL == url {
			return record.ShortURL, nil
		} else if searchByType == ShortURLType && record.ShortURL == url {
			return record.OriginalURL, nil
		}
	}
	return "", ErrURLStorageDataNotFound
}

func (s *FileURLStorage) Set(origURL, shortURL string) error {
	s.logger.Debug("Setting url to storage",
		zap.String("origURL", origURL),
		zap.String("shortURL", shortURL),
	)
	record := fileRecord{
		UUID:        s.uuidMgr.next(),
		ShortURL:    shortURL,
		OriginalURL: origURL,
	}
	*s.records = append(*s.records, record)
	if err := s.persistToFile(record); err != nil {
		return err
	}
	return nil
}

func (s *FileURLStorage) BatchSet(binds *[]URLBind) error {
	s.logger.Debug("Setting batch of url bindings to storage",
		zap.Int("count", len(*binds)),
	)
	for _, b := range *binds {
		if err := s.Set(b.OrigURL, b.ShortID); err != nil {
			return err
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
		s.logger.Warn("Can't restore from requested file, trying default", zap.Error(err))
		if err := s.restoreFromFile(true); err != nil {
			s.logger.Error("Failed to restore from default file", zap.Error(err))
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	records, err := s.fileScnr.scan(file)
	if err != nil && !useDefault {
		s.logger.Warn("Can't scan data from requested file, trying default", zap.Error(err))
		s.fileMgr.close()
		if err := s.restoreFromFile(true); err != nil {
			s.fileMgr.close()
			s.logger.Error("Failed to scan data from default file", zap.Error(err))
			return err
		}
		return nil
	} else if err != nil {
		s.fileMgr.close()
		return err
	}

	s.records = records
	return nil
}

var (
	ErrURLStorageDataNotFound = errors.New("no data in the storage for the requested url")
)
