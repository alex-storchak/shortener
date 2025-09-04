package service

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stubDecoder struct {
	retReq model.ShortenRequest
	retErr error
}

func (s *stubDecoder) Decode(_ io.Reader) (model.ShortenRequest, error) {
	return s.retReq, s.retErr
}

type stubCore struct {
	retShortURL string
	retErr      error
}

func (s *stubCore) Shorten(_ string) (string, string, error) {
	return s.retShortURL, "", s.retErr
}

func TestAPIShortenService_Shorten(t *testing.T) {
	tests := []struct {
		name         string
		body         []byte
		decoderReq   model.ShortenRequest
		decoderErr   error
		coreShortURL string
		coreErr      error
		wantResp     *model.ShortenResponse
		wantErr      bool
		wantErrIs    error
	}{
		{
			name:       "returns ErrJSONDecode when decoder fails",
			body:       []byte("{bad json}"),
			decoderErr: errors.New("decode error"),
			wantErr:    true,
			wantErrIs:  ErrJSONDecode,
		},
		{
			name:       "maps ErrEmptyInputURL to ErrEmptyURL",
			body:       []byte("{}"),
			decoderReq: model.ShortenRequest{OrigURL: ""},
			coreErr:    ErrEmptyInputURL,
			wantErr:    true,
			wantErrIs:  ErrEmptyURL,
		},
		{
			name:       "returns unexpected core error",
			body:       []byte("{}"),
			decoderReq: model.ShortenRequest{OrigURL: "https://example.com"},
			coreErr:    errors.New("random error"),
			wantErr:    true,
		},
		{
			name:         "success returns short url in response",
			body:         []byte("{}"),
			decoderReq:   model.ShortenRequest{OrigURL: "https://example.com"},
			coreShortURL: "https://short.host/abcde",
			wantResp:     &model.ShortenResponse{ShortURL: "https://short.host/abcde"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := &stubDecoder{tt.decoderReq, tt.decoderErr}
			core := &stubCore{tt.coreShortURL, tt.coreErr}
			srv := NewAPIShortenService(core, dec, zap.NewNop())

			var r io.Reader = bytes.NewReader(tt.body)
			resp, err := srv.Shorten(r)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				}
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantResp, resp)
		})
	}
}
