package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type idGeneratorStub struct {
	generateMethodShouldFail bool
}

func newIDGeneratorStub(generateMethodShouldFail bool) *idGeneratorStub {
	return &idGeneratorStub{
		generateMethodShouldFail: generateMethodShouldFail,
	}
}

func (d *idGeneratorStub) Generate() (string, error) {
	if d.generateMethodShouldFail {
		return "", errors.New("generate method should fail")
	}
	return "abcde", nil
}

type urlStorageStub struct {
	setMethodShouldFail bool
	storage             map[string]string
}

func newURLStorageStub(setMethodShouldFail bool) *urlStorageStub {
	return &urlStorageStub{
		setMethodShouldFail: setMethodShouldFail,
		storage: map[string]string{
			"http://existing.com": "abcde",
		},
	}
}

func newShortURLStorageStub(setMethodShouldFail bool) *urlStorageStub {
	return &urlStorageStub{
		setMethodShouldFail: setMethodShouldFail,
		storage: map[string]string{
			"abcde": "http://existing.com",
		},
	}
}

func (d *urlStorageStub) Has(url string) bool {
	_, ok := d.storage[url]
	return ok
}

func (d *urlStorageStub) Get(url string) (string, error) {
	if targetURL, ok := d.storage[url]; ok {
		return targetURL, nil
	} else {
		return "", errors.New("no data in the storage for the requested url")
	}
}

func (d *urlStorageStub) Set(urlKey, urlValue string) error {
	if d.setMethodShouldFail {
		return errors.New("set method should fail")
	}
	return nil
}

func TestShortener_Shorten(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name                      string
		args                      args
		idGeneratorShouldFail     bool
		urlStorageShouldFail      bool
		shortURLStorageShouldFail bool
		err                       error
		want                      string
		wantErr                   bool
	}{
		{
			name: "returns new short id if storage is empty",
			args: args{
				url: "https://non-existing.com",
			},
			err:     nil,
			want:    "abcde",
			wantErr: false,
		},
		{
			name: "returns short id from storage if exists",
			args: args{
				url: "http://existing.com",
			},
			err:     nil,
			want:    "abcde",
			wantErr: false,
		},
		{
			name: "returns error if generation of short id is failed",
			args: args{
				url: "https://non-existing.com",
			},
			idGeneratorShouldFail: true,
			err:                   ErrShortenerGenerationShortIDFailed,
			want:                  "",
			wantErr:               true,
		},
		{
			name: "returns error if failed to set binding in urlStorage",
			args: args{
				url: "http://non-existing.com",
			},
			urlStorageShouldFail: true,
			err:                  ErrShortenerSetBindingURLStorageFailed,
			want:                 "",
			wantErr:              true,
		},
		{
			name: "returns error if failed to set binding in shortURLStorage",
			args: args{
				url: "http://non-existing.com",
			},
			shortURLStorageShouldFail: true,
			err:                       ErrShortenerSetBindingShortURLStorageFailed,
			want:                      "",
			wantErr:                   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Shortener{
				urlStorage:      newURLStorageStub(tt.urlStorageShouldFail),
				shortURLStorage: newShortURLStorageStub(tt.shortURLStorageShouldFail),
				generator:       newIDGeneratorStub(tt.idGeneratorShouldFail),
			}

			got, err := s.Shorten(tt.args.url)

			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestShortener_Extract(t *testing.T) {
	type args struct {
		shortID string
	}
	tests := []struct {
		name                      string
		args                      args
		shortURLStorageShouldFail bool
		err                       error
		want                      string
		wantErr                   bool
	}{
		{
			name: "extract url from storage if exists",
			args: args{
				shortID: "abcde",
			},
			err:     nil,
			want:    "http://existing.com",
			wantErr: false,
		},
		{
			name: "returns error if generation of short id is failed",
			args: args{
				shortID: "non-existing",
			},
			shortURLStorageShouldFail: true,
			err:                       ErrShortenerShortIDNotFound,
			want:                      "",
			wantErr:                   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Shortener{
				urlStorage:      newURLStorageStub(false),
				shortURLStorage: newShortURLStorageStub(tt.shortURLStorageShouldFail),
				generator:       newIDGeneratorStub(false),
			}

			got, err := s.Extract(tt.args.shortID)

			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
