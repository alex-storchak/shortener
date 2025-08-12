package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// todo: перделать на использование стабов
type idGeneratorMock struct {
	mock.Mock
}

func (m *idGeneratorMock) Generate() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

type URLStorageMock struct {
	mock.Mock
}

func (m *URLStorageMock) Has(shortID string) bool {
	args := m.Called(shortID)
	return args.Bool(0)
}

func (m *URLStorageMock) Get(shortID string) (string, error) {
	args := m.Called(shortID)
	return args.String(0), args.Error(1)
}

func (m *URLStorageMock) Set(shortID, url string) error {
	args := m.Called(shortID, url)
	return args.Error(0)
}

func TestShortener_Shorten(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name     string
		args     args
		mockFunc func(idm *idGeneratorMock, usm *URLStorageMock, susm *URLStorageMock)
		err      error
		want     string
		wantErr  bool
	}{
		{
			name: "returns new short id if storage is empty",
			args: args{
				url: "https://non-existing.com",
			},
			mockFunc: func(idm *idGeneratorMock, usm *URLStorageMock, susm *URLStorageMock) {
				usm.On("Has", "https://non-existing.com").Return(false)
				idm.On("Generate").Return("abcde", nil)
				usm.On("Set", "https://non-existing.com", "abcde").Return(nil)
				susm.On("Set", "abcde", "https://non-existing.com").Return(nil)
			},
			err:     nil,
			want:    "abcde",
			wantErr: false,
		},
		{
			name: "returns short id from storage if exists",
			args: args{
				url: "https://existing.com",
			},
			mockFunc: func(idm *idGeneratorMock, usm *URLStorageMock, susm *URLStorageMock) {
				usm.On("Has", "https://existing.com").Return(true)
				usm.On("Get", "https://existing.com").Return("abcde", nil)
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
			mockFunc: func(idm *idGeneratorMock, usm *URLStorageMock, susm *URLStorageMock) {
				usm.On("Has", "https://non-existing.com").Return(false)
				idm.On("Generate").Return("", errors.New("failed"))
			},
			err:     ErrShortenerGenerationShortIDFailed,
			want:    "",
			wantErr: true,
		},
		{
			name: "returns error if failed to set binding in urlStorage",
			args: args{
				url: "https://non-existing.com",
			},
			mockFunc: func(idm *idGeneratorMock, usm *URLStorageMock, susm *URLStorageMock) {
				usm.On("Has", "https://non-existing.com").Return(false)
				idm.On("Generate").Return("abcde", nil)
				usm.On("Set", "https://non-existing.com", "abcde").Return(errors.New("failed"))
			},
			err:     ErrShortenerSetBindingURLStorageFailed,
			want:    "",
			wantErr: true,
		},
		{
			name: "returns error if failed to set binding in shortURLStorage",
			args: args{
				url: "https://non-existing.com",
			},
			mockFunc: func(idm *idGeneratorMock, usm *URLStorageMock, susm *URLStorageMock) {
				usm.On("Has", "https://non-existing.com").Return(false)
				idm.On("Generate").Return("abcde", nil)
				usm.On("Set", "https://non-existing.com", "abcde").Return(nil)
				susm.On("Set", "abcde", "https://non-existing.com").Return(errors.New("failed"))
			},
			err:     ErrShortenerSetBindingShortURLStorageFailed,
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Shortener{
				urlStorage:      &URLStorageMock{},
				shortURLStorage: &URLStorageMock{},
				generator:       &idGeneratorMock{},
			}
			tt.mockFunc(
				s.generator.(*idGeneratorMock),
				s.urlStorage.(*URLStorageMock),
				s.shortURLStorage.(*URLStorageMock),
			)

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
		name     string
		args     args
		mockFunc func(idm *idGeneratorMock, usm *URLStorageMock, susm *URLStorageMock)
		err      error
		want     string
		wantErr  bool
	}{
		{
			name: "extract url from storage if exists",
			args: args{
				shortID: "abcde",
			},
			mockFunc: func(idm *idGeneratorMock, usm *URLStorageMock, susm *URLStorageMock) {
				susm.On("Has", "abcde").Return(true)
				susm.On("Get", "abcde").Return("https://existing.com", nil)
			},
			err:     nil,
			want:    "https://existing.com",
			wantErr: false,
		},
		{
			name: "returns error if generation of short id is failed",
			args: args{
				shortID: "non-existing",
			},
			mockFunc: func(idm *idGeneratorMock, usm *URLStorageMock, susm *URLStorageMock) {
				susm.On("Has", "non-existing").Return(false)
			},
			err:     ErrShortenerShortIDNotFound,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Shortener{
				urlStorage:      &URLStorageMock{},
				shortURLStorage: &URLStorageMock{},
				generator:       &idGeneratorMock{},
			}
			tt.mockFunc(
				s.generator.(*idGeneratorMock),
				s.urlStorage.(*URLStorageMock),
				s.shortURLStorage.(*URLStorageMock),
			)

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
