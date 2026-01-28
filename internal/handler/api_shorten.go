package handler

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/codec"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

// APIShortenProcessor defines the interface for processing API shorten requests.
type APIShortenProcessor interface {
	Process(ctx context.Context, req model.ShortenRequest) (*model.ShortenResponse, error)
}

// HandleAPIShorten creates an HTTP handler for the API URL shortening endpoint.
// It validates content type, decodes JSON request, processes the shortening operation,
// and publishes audit events for successful shortenings.
//
// Returns:
// - 400 Bad Request for invalid content type or malformed JSON
// - 400 Bad Request for empty input URL
// - 409 Conflict when URL already exists (returns existing short URL)
// - 201 Created for successful shortening
// - 500 Internal Server Error for processing failures
func HandleAPIShorten(p APIShortenProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if err := validateContentType(ct, "application/json"); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var req model.ShortenRequest
		if err := codec.EasyJSONDecode(r, &req); err != nil {
			l.Debug("decode json request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resBody, err := p.Process(r.Context(), req)
		if errors.Is(err, service.ErrEmptyInputURL) {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrURLAlreadyExists) {
			if err = codec.EasyJSONEncode(w, http.StatusConflict, resBody); err != nil {
				l.Error("conflict. encode json response", zap.Error(err))
			}
			return
		} else if err != nil {
			l.Error("failed to shorten", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = codec.EasyJSONEncode(w, http.StatusCreated, resBody); err != nil {
			l.Error("created. encode json response", zap.Error(err))
			return
		}
	}
}
