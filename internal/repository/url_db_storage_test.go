package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type dbManagerStub struct {
	retByOrig  string
	errByOrig  error
	retByShort string
	errByShort error
	persistErr error
}

func (s dbManagerStub) Close() error {
	return nil
}

func (s dbManagerStub) Ping(_ context.Context) error {
	return nil
}

func (s dbManagerStub) GetByOriginalURL(_ context.Context, origURL string) (*model.URLStorageRecord, error) {
	return &model.URLStorageRecord{
		OrigURL: origURL,
		ShortID: s.retByOrig,
	}, s.errByOrig
}

func (s dbManagerStub) GetByShortID(_ context.Context, shortID string) (*model.URLStorageRecord, error) {
	return &model.URLStorageRecord{
		OrigURL: s.retByShort,
		ShortID: shortID,
	}, s.errByShort
}

func (s dbManagerStub) Persist(_ context.Context, _ *model.URLStorageRecord) error {
	return s.persistErr
}

func (s dbManagerStub) PersistBatch(_ context.Context, _ []*model.URLStorageRecord) error {
	return s.persistErr
}

func (s dbManagerStub) GetByUserUUID(_ context.Context, _ string) ([]*model.URLStorageRecord, error) {
	return nil, nil
}

func (s dbManagerStub) DeleteBatch(_ context.Context, _ model.URLDeleteBatch) error {
	return nil
}

func TestDBURLStorage_Get(t *testing.T) {
	lgr := zap.NewNop()
	unexpected := errors.New("random error")
	var notFound *DataNotFoundError

	tests := []struct {
		name       string
		searchType string
		stub       dbManagerStub
		inputURL   string
		wantURL    model.URLStorageRecord
		wantErr    bool
		wantErrIs  error
		wantErrAs  any
	}{
		{
			name:       "success by original",
			searchType: OrigURLType,
			stub:       dbManagerStub{retByOrig: "abcde"},
			inputURL:   "https://example.com",
			wantURL:    model.URLStorageRecord{OrigURL: "https://example.com", ShortID: "abcde"},
		},
		{
			name:       "success by short",
			searchType: ShortURLType,
			stub:       dbManagerStub{retByShort: "https://example.com"},
			inputURL:   "abcde",
			wantURL:    model.URLStorageRecord{OrigURL: "https://example.com", ShortID: "abcde"},
		},
		{
			name:       "returns not found error when not found by original url",
			searchType: OrigURLType,
			stub:       dbManagerStub{errByOrig: ErrDataNotFoundInDB},
			inputURL:   "https://missing.com",
			wantErr:    true,
			wantErrAs:  &notFound,
		},
		{
			name:       "returns not found error when not found by short url",
			searchType: ShortURLType,
			stub:       dbManagerStub{errByShort: ErrDataNotFoundInDB},
			inputURL:   "missingShort",
			wantErr:    true,
			wantErrAs:  &notFound,
		},
		{
			name:       "return unexpected error by original",
			searchType: OrigURLType,
			stub:       dbManagerStub{errByOrig: unexpected},
			inputURL:   "https://example.com",
			wantErr:    true,
			wantErrIs:  unexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewDBURLStorage(lgr, tt.stub)
			got, err := st.Get(tt.inputURL, tt.searchType)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				}
				if tt.wantErrAs != nil {
					require.ErrorAs(t, err, tt.wantErrAs)
				}
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantURL.OrigURL, got.OrigURL)
			assert.Equal(t, tt.wantURL.ShortID, got.ShortID)
		})
	}
}

func TestDBURLStorage_Set(t *testing.T) {
	lgr := zap.NewNop()
	insertErr := errors.New("insert failed")

	tests := []struct {
		name      string
		stub      dbManagerStub
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "success",
			stub: dbManagerStub{persistErr: nil},
		},
		{
			name:      "returns persist error",
			stub:      dbManagerStub{persistErr: insertErr},
			wantErr:   true,
			wantErrIs: insertErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewDBURLStorage(lgr, tt.stub)
			err := st.Set(&model.URLStorageRecord{OrigURL: "https://example.com", ShortID: "abcde"})
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestDBURLStorage_BatchSet(t *testing.T) {
	lgr := zap.NewNop()
	insertErr := errors.New("batch insert failed")

	tests := []struct {
		name      string
		stub      dbManagerStub
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "success",
			stub: dbManagerStub{persistErr: nil},
		},
		{
			name:      "returns persist error",
			stub:      dbManagerStub{persistErr: insertErr},
			wantErr:   true,
			wantErrIs: insertErr,
		},
	}

	binds := []*model.URLStorageRecord{
		{OrigURL: "https://a.com", ShortID: "abc"},
		{OrigURL: "https://b.com", ShortID: "def"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewDBURLStorage(lgr, tt.stub)
			err := st.BatchSet(binds)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}
