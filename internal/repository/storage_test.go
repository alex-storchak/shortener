package repository

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFileURLStorage(t *testing.T) {
	tmpDir := t.TempDir()
	testDBFile, err := os.CreateTemp(tmpDir, "file_db*.txt")
	require.NoError(t, err)
	defer os.Remove(testDBFile.Name())

	testRecord := FileURLStorageRecord{
		UUID:        1,
		ShortURL:    "abcde",
		OriginalURL: "https://example.com",
	}
	data, err := json.Marshal(testRecord)
	require.NoError(t, err)
	dataWithNewline := append(data, '\n')
	os.WriteFile(testDBFile.Name(), dataWithNewline, 0666)

	tests := []struct {
		name            string
		fileStoragePath string
		hasRecord       bool
		wantShortURL    string
		wantOrigURL     string
	}{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewFileURLStorage(tt.fileStoragePath, zap.NewNop())
			require.NoError(t, err)

			if tt.hasRecord {
				assert.True(t, storage.Has(tt.wantShortURL, ShortURLType))
				assert.True(t, storage.Has(tt.wantOrigURL, OrigURLType))
			} else {
				assert.False(t, storage.Has(tt.wantShortURL, ShortURLType))
				assert.False(t, storage.Has(tt.wantOrigURL, OrigURLType))

				err := storage.Set(tt.wantOrigURL, tt.wantShortURL)
				require.NoError(t, err)

				assert.True(t, storage.Has(tt.wantShortURL, ShortURLType))
				assert.True(t, storage.Has(tt.wantOrigURL, OrigURLType))

				origURL, err := storage.Get(tt.wantShortURL, ShortURLType)
				require.NoError(t, err)
				assert.Equal(t, tt.wantOrigURL, origURL)

				shortURL, err := storage.Get(tt.wantOrigURL, OrigURLType)
				require.NoError(t, err)
				assert.Equal(t, tt.wantShortURL, shortURL)

				newStorage, err := NewFileURLStorage(tt.fileStoragePath, zap.NewNop())
				require.NoError(t, err)
				assert.True(t, newStorage.Has(tt.wantShortURL, ShortURLType))
				assert.True(t, newStorage.Has(tt.wantOrigURL, OrigURLType))
			}
		})
	}
}
