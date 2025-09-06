package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewMapURLStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MapURLStorage
	}{
		{
			name: "default creation",
			want: &MapURLStorage{
				storage: make(map[string]string),
				logger:  zap.NewNop(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMapURLStorage(zap.NewNop()))
		})
	}
}

func TestMapURLStorage_Has(t *testing.T) {
	storage := MapURLStorage{
		storage: map[string]string{
			"https://existing.com": "http://localhost:8080/EwHXdJfB",
		},
		logger: zap.NewNop(),
	}

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			"has url",
			"https://existing.com",
			true,
		},
		{
			"no url",
			"https://not-existing.com",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storage.Has(tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapURLStorage_Get(t *testing.T) {
	storage := MapURLStorage{
		storage: map[string]string{
			"https://existing.com": "http://localhost:8080/EwHXdJfB",
		},
		logger: zap.NewNop(),
	}

	tests := []struct {
		name    string
		url     string
		err     error
		want    string
		wantErr bool
	}{
		{
			"has url",
			"https://existing.com",
			nil,
			"http://localhost:8080/EwHXdJfB",
			false,
		},
		{
			"no url error",
			"https://not-existing.com",
			ErrURLStorageDataNotFound,
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.Get(tt.url)
			if tt.wantErr {
				assert.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMapURLStorage_Set(t *testing.T) {
	storage := MapURLStorage{
		storage: make(map[string]string),
		logger:  zap.NewNop(),
	}

	tests := []struct {
		name     string
		url      string
		shortURL string
	}{
		{
			"set url",
			"https://existing.com",
			"http://localhost:8080/EwHXdJfB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.Set(tt.url, tt.shortURL)
			require.NoError(t, err)

			got, ok := storage.storage[tt.url]
			require.True(t, ok)
			assert.Equal(t, tt.shortURL, got)
		})
	}
}

func TestMapURLStorage_allMethods(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		shortURL string
	}{
		{
			name:     "set url",
			url:      "https://existing.com",
			shortURL: "http://localhost:8080/EwHXdJfB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMapURLStorage(zap.NewNop())
			err := storage.Set(tt.url, tt.shortURL)
			require.NoError(t, err)
			got, err := storage.Get(tt.url)
			require.NoError(t, err)
			assert.Equal(t, tt.shortURL, got)
		})
	}
}
