package repository

import (
	"context"
	"errors"
	"testing"

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

func (s dbManagerStub) GetByOriginalURL(_ context.Context, _ string) (string, error) {
	return s.retByOrig, s.errByOrig
}

func (s dbManagerStub) GetByShortID(_ context.Context, _ string) (string, error) {
	return s.retByShort, s.errByShort
}

func (s dbManagerStub) Persist(_ context.Context, _, _ string) error {
	return s.persistErr
}

func (s dbManagerStub) PersistBatch(_ context.Context, _ *[]URLBind) error {
	return s.persistErr
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
		wantURL    string
		wantErr    bool
		wantErrIs  error
		wantErrAs  any
	}{
		{
			name:       "success by original",
			searchType: OrigURLType,
			stub:       dbManagerStub{retByOrig: "abcde"},
			inputURL:   "https://example.com",
			wantURL:    "abcde",
		},
		{
			name:       "success by short",
			searchType: ShortURLType,
			stub:       dbManagerStub{retByShort: "https://example.com"},
			inputURL:   "abcde",
			wantURL:    "https://example.com",
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
				assert.Equal(t, "", got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, got)
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
			err := st.Set("https://example.com", "abcde")
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

	binds := &[]URLBind{
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
