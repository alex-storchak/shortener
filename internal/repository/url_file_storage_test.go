package repository

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/alex-storchak/shortener/internal/file"
	"github.com/alex-storchak/shortener/internal/model"
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

	userUUID := "userUUID"

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
			fm := file.NewManager(tt.fileStoragePath, tt.dfltStoragePath, lgr)
			frp := URLFileRecordParser{}
			fs := NewFileScanner(lgr, frp)
			storage, err := NewFileURLStorage(lgr, fm, fs)
			require.NoError(t, err)

			if tt.hasRecord {
				assertStorageHasURL(t, tt, storage)
				return
			}

			assertStorageDoesNotHaveURL(t, tt, storage)

			err = storage.Set(&model.URLStorageRecord{OrigURL: tt.wantOrigURL, ShortID: tt.wantShortURL, UserUUID: userUUID})
			require.NoError(t, err)

			assertStorageHasURL(t, tt, storage)

			fm = file.NewManager(tt.fileStoragePath, tt.dfltStoragePath, lgr)
			newStorage, err := NewFileURLStorage(lgr, fm, fs)
			require.NoError(t, err)
			assertStorageHasURL(t, tt, newStorage)
		})
	}
}

func assertStorageHasURL(t *testing.T, tt testCaseData, storage URLStorage) {
	ou, err := storage.Get(tt.wantShortURL, ShortURLType)
	require.NoError(t, err)
	assert.Equal(t, tt.wantOrigURL, ou.OrigURL)

	su, err := storage.Get(tt.wantOrigURL, OrigURLType)
	require.NoError(t, err)
	assert.Equal(t, tt.wantShortURL, su.ShortID)
}

func assertStorageDoesNotHaveURL(t *testing.T, tt testCaseData, storage URLStorage) {
	_, err := storage.Get(tt.wantShortURL, ShortURLType)
	var nfErrShort, nfErrOrig *DataNotFoundError
	require.ErrorAs(t, err, &nfErrShort)
	_, err = storage.Get(tt.wantOrigURL, OrigURLType)
	require.ErrorAs(t, err, &nfErrOrig)
}

func createTmpStorageFile(t *testing.T) *os.File {
	tmpDir := t.TempDir()
	testDBFile, err := os.CreateTemp(tmpDir, "file_db*.txt")
	require.NoError(t, err)
	return testDBFile
}

func fillStorageFile(t *testing.T, testDBFile *os.File) {
	testRecord := model.URLStorageRecord{
		ShortID:  "abcde",
		OrigURL:  "https://example.com",
		UserUUID: "userUUID",
	}
	data, err := json.Marshal(testRecord)
	require.NoError(t, err)
	dataWithNewline := append(data, '\n')
	if err := os.WriteFile(testDBFile.Name(), dataWithNewline, 0666); err != nil {
		t.Fatalf("failed to write test record to storage file: %v", err)
	}
}

func fillBadStorageFile(t *testing.T, testDBFile *os.File) {
	data := []byte("foo")
	if err := os.WriteFile(testDBFile.Name(), data, 0666); err != nil {
		t.Fatalf("failed to write bad test record to bad storage file: %v", err)
	}
}
