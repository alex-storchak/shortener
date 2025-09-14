package repository

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type testCaseData struct {
	name            string
	fileStoragePath string
	dfltStoragePath string
	hasRecord       bool
	wantShortURL    string
	wantOrigURL     string
}

func TestFileURLStorage(t *testing.T) {
	testDBFile := createTmpStorageFile(t)
	defer os.Remove(testDBFile.Name())

	badTestDBFile := createTmpStorageFile(t)
	defer os.Remove(badTestDBFile.Name())

	tests := []testCaseData{
		{
			name:            "return preload record from storage file",
			fileStoragePath: testDBFile.Name(),
			hasRecord:       true,
			wantShortURL:    "abcde",
			wantOrigURL:     "https://example.com",
		},
		{
			name:            "return non-existing url from storage after set",
			fileStoragePath: testDBFile.Name(),
			hasRecord:       false,
			wantShortURL:    "some_non_existing",
			wantOrigURL:     "https://non-existing.com",
		},
		{
			name:            "recovery storage because of bad file path return non-existing url from storage after set",
			fileStoragePath: "/non-existing/path",
			dfltStoragePath: testDBFile.Name(),
			hasRecord:       false,
			wantShortURL:    "some_non_existing",
			wantOrigURL:     "https://non-existing.com",
		},
		{
			name:            "recovery storage because of bad file content return non-existing url from storage after set",
			fileStoragePath: badTestDBFile.Name(),
			dfltStoragePath: testDBFile.Name(),
			hasRecord:       false,
			wantShortURL:    "some_non_existing",
			wantOrigURL:     "https://non-existing.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fillStorageFile(t, testDBFile)
			fillBadStorageFile(t, badTestDBFile)

			lgr := zap.NewNop()
			fm := NewFileManager(tt.fileStoragePath, tt.dfltStoragePath, lgr)
			frp := FileRecordParser{}
			fs := NewFileScanner(lgr, frp)
			um := NewUUIDManager(lgr)
			storage, err := NewFileURLStorage(lgr, fm, fs, um)
			require.NoError(t, err)

			if tt.hasRecord {
				assertStorageHasURL(t, tt, storage)
				return
			}

			assertStorageDoesNotHaveURL(t, tt, storage)

			err = storage.Set(tt.wantOrigURL, tt.wantShortURL)
			require.NoError(t, err)

			assertStorageHasURL(t, tt, storage)

			fm = NewFileManager(tt.fileStoragePath, tt.dfltStoragePath, lgr)
			um = NewUUIDManager(lgr)
			newStorage, err := NewFileURLStorage(lgr, fm, fs, um)
			require.NoError(t, err)
			assertStorageHasURL(t, tt, newStorage)
		})
	}
}

func assertStorageHasURL(t *testing.T, tt testCaseData, storage URLStorage) {
	origURL, err := storage.Get(tt.wantShortURL, ShortURLType)
	require.NoError(t, err)
	assert.Equal(t, tt.wantOrigURL, origURL)

	shortURL, err := storage.Get(tt.wantOrigURL, OrigURLType)
	require.NoError(t, err)
	assert.Equal(t, tt.wantShortURL, shortURL)
}

func assertStorageDoesNotHaveURL(t *testing.T, tt testCaseData, storage URLStorage) {
	_, err := storage.Get(tt.wantShortURL, ShortURLType)
	require.ErrorIs(t, err, ErrURLStorageDataNotFound)
	_, err = storage.Get(tt.wantOrigURL, OrigURLType)
	require.ErrorIs(t, err, ErrURLStorageDataNotFound)
}

func createTmpStorageFile(t *testing.T) *os.File {
	tmpDir := t.TempDir()
	testDBFile, err := os.CreateTemp(tmpDir, "file_db*.txt")
	require.NoError(t, err)
	return testDBFile
}

func fillStorageFile(t *testing.T, testDBFile *os.File) {
	testRecord := fileRecord{
		UUID:        1,
		ShortURL:    "abcde",
		OriginalURL: "https://example.com",
	}
	data, err := json.Marshal(testRecord)
	require.NoError(t, err)
	dataWithNewline := append(data, '\n')
	_ = os.WriteFile(testDBFile.Name(), dataWithNewline, 0666)
}

func fillBadStorageFile(_ *testing.T, testDBFile *os.File) {
	data := []byte("foo")
	_ = os.WriteFile(testDBFile.Name(), data, 0666)
}
