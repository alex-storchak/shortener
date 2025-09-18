package service

import (
	"errors"
	"testing"

	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
	setMethodShouldFail      bool
	setBatchMethodShouldFail bool
	storage                  []urlStorageStubRecord
}

type urlStorageStubRecord struct {
	shortURL string
	origURL  string
}

func newURLStorageStub(
	setMethodShouldFail bool,
	setBatchMethodShouldFail bool,
) *urlStorageStub {
	return &urlStorageStub{
		setMethodShouldFail:      setMethodShouldFail,
		setBatchMethodShouldFail: setBatchMethodShouldFail,
		storage: []urlStorageStubRecord{
			{
				shortURL: "abcde",
				origURL:  "http://existing.com",
			},
		},
	}
}

func (d *urlStorageStub) Get(url, searchByType string) (string, error) {
	if searchByType == repository.OrigURLType && d.storage[0].origURL == url {
		return d.storage[0].shortURL, nil
	} else if searchByType == repository.ShortURLType && d.storage[0].shortURL == url {
		return d.storage[0].origURL, nil
	}
	return "", repository.ErrURLStorageDataNotFound
}

func (d *urlStorageStub) Set(_, _ string) error {
	if d.setMethodShouldFail {
		return errors.New("set method should fail")
	}
	return nil
}

func (d *urlStorageStub) BatchSet(_ *[]repository.URLBind) error {
	if d.setBatchMethodShouldFail {
		return errors.New("set batch method should fail")
	}
	return nil
}
func TestShortener_Shorten(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name                  string
		args                  args
		idGeneratorShouldFail bool
		urlStorageShouldFail  bool
		err                   error
		want                  string
		wantErr               bool
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
			name: "returns short id from storage if exists and ErrURLAlreadyExists",
			args: args{
				url: "http://existing.com",
			},
			err:     ErrURLAlreadyExists,
			want:    "abcde",
			wantErr: true,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Shortener{
				urlStorage: newURLStorageStub(tt.urlStorageShouldFail, false),
				generator:  newIDGeneratorStub(tt.idGeneratorShouldFail),
				logger:     zap.NewNop(),
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
		name    string
		args    args
		err     error
		want    string
		wantErr bool
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
			name: "returns error if extracting by short id is failed",
			args: args{
				shortID: "non-existing",
			},
			err:     repository.ErrURLStorageDataNotFound,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Shortener{
				urlStorage: newURLStorageStub(false, false),
				generator:  newIDGeneratorStub(false),
				logger:     zap.NewNop(),
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

func TestShortener_ShortenBatch(t *testing.T) {
	tests := []struct {
		name                  string
		urls                  []string
		idGeneratorShouldFail bool
		batchSetShouldFail    bool
		want                  []string
		wantErr               bool
		err                   error
	}{
		{
			name: "success returns ids for existing and new",
			urls: []string{"http://existing.com", "https://non-existing.com"},
			want: []string{"abcde", "abcde"},
		},
		{
			name:    "returns ErrEmptyInputURL when any url is empty",
			urls:    []string{""},
			wantErr: true,
			err:     ErrEmptyInputURL,
		},
		{
			name:                  "returns ErrShortenerGenerationShortIDFailed when generator fails",
			urls:                  []string{"https://non-existing.com"},
			idGeneratorShouldFail: true,
			wantErr:               true,
			err:                   ErrShortenerGenerationShortIDFailed,
		},
		{
			name:               "returns ErrShortenerSetBindingURLStorageFailed when BatchSet fails",
			urls:               []string{"https://non-existing.com"},
			batchSetShouldFail: true,
			wantErr:            true,
			err:                ErrShortenerSetBindingURLStorageFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := newURLStorageStub(false, tt.batchSetShouldFail)
			s := Shortener{
				urlStorage: us,
				generator:  newIDGeneratorStub(tt.idGeneratorShouldFail),
				logger:     zap.NewNop(),
			}

			got, err := s.ShortenBatch(&tt.urls)

			if !tt.wantErr {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want, *got)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
				assert.Nil(t, got)
			}
		})
	}
}
