package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/model"
	repo "github.com/alex-storchak/shortener/internal/repository"
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
	storage                  []model.URLStorageRecord
}

func newURLStorageStub(
	setMethodShouldFail bool,
	setBatchMethodShouldFail bool,
) *urlStorageStub {
	return &urlStorageStub{
		setMethodShouldFail:      setMethodShouldFail,
		setBatchMethodShouldFail: setBatchMethodShouldFail,
		storage: []model.URLStorageRecord{
			{
				OrigURL: "http://existing.com",
				ShortID: "abcde",
			},
		},
	}
}

func (d *urlStorageStub) Close() error {
	return nil
}

func (d *urlStorageStub) Ping(_ context.Context) error {
	return nil
}

func (d *urlStorageStub) Get(_ context.Context, url, searchByType string) (*model.URLStorageRecord, error) {
	if searchByType == repo.OrigURLType && d.storage[0].OrigURL == url {
		return &d.storage[0], nil
	} else if searchByType == repo.ShortURLType && d.storage[0].ShortID == url {
		return &d.storage[0], nil
	}
	return nil, repo.NewDataNotFoundError(nil)
}

func (d *urlStorageStub) Set(_ context.Context, _, _, _ string) error {
	if d.setMethodShouldFail {
		return errors.New("set method should fail")
	}
	return nil
}

func (d *urlStorageStub) BatchSet(_ context.Context, _ []model.URLStorageRecord) error {
	if d.setBatchMethodShouldFail {
		return errors.New("set batch method should fail")
	}
	return nil
}

func (d *urlStorageStub) GetByUserUUID(_ context.Context, _ string) ([]*model.URLStorageRecord, error) {
	return nil, nil
}

func (d *urlStorageStub) DeleteBatch(_ context.Context, _ model.URLDeleteBatch) error {
	return nil
}

func TestShortener_Shorten(t *testing.T) {
	type args struct {
		url string
	}
	userUUID := "userUUID"

	tests := []struct {
		name              string
		args              args
		idGenShouldFail   bool
		storageShouldFail bool
		err               error
		want              string
		wantErr           bool
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
			idGenShouldFail: true,
			want:            "",
			wantErr:         true,
		},
		{
			name: "returns error if failed to set binding in urlStorage",
			args: args{
				url: "http://non-existing.com",
			},
			storageShouldFail: true,
			want:              "",
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Shortener{
				urlStorage: newURLStorageStub(tt.storageShouldFail, false),
				generator:  newIDGeneratorStub(tt.idGenShouldFail),
				logger:     zap.NewNop(),
			}

			got, err := s.Shorten(t.Context(), userUUID, tt.args.url)

			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				require.Error(t, err)
				if tt.err != nil {
					assert.ErrorIs(t, err, tt.err)
				}
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestShortener_Extract(t *testing.T) {
	var nfErr *repo.DataNotFoundError

	type args struct {
		shortID string
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantErr   bool
		wantErrAs any
	}{
		{
			name: "extract url from storage if exists",
			args: args{
				shortID: "abcde",
			},
			want:    "http://existing.com",
			wantErr: false,
		},
		{
			name: "returns error if extracting by short id is failed",
			args: args{
				shortID: "non-existing",
			},
			want:      "",
			wantErr:   true,
			wantErrAs: &nfErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Shortener{
				urlStorage: newURLStorageStub(false, false),
				generator:  newIDGeneratorStub(false),
				logger:     zap.NewNop(),
			}

			got, err := s.Extract(t.Context(), tt.args.shortID)

			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				require.Error(t, err)
				assert.ErrorAs(t, err, tt.wantErrAs)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestShortener_ShortenBatch(t *testing.T) {
	userUUID := "userUUID"

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
		},
		{
			name:               "returns ErrShortenerSetBindingURLStorageFailed when BatchSet fails",
			urls:               []string{"https://non-existing.com"},
			batchSetShouldFail: true,
			wantErr:            true,
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

			got, err := s.ShortenBatch(t.Context(), userUUID, tt.urls)

			if !tt.wantErr {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			} else {
				require.Error(t, err)
				if tt.err != nil {
					assert.ErrorIs(t, err, tt.err)
				}
				assert.Nil(t, got)
			}
		})
	}
}
