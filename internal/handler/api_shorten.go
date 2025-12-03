package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/codec"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

type APIShortenProcessor interface {
	Process(ctx context.Context, req model.ShortenRequest) (*model.ShortenResponse, string, error)
}

func handleAPIShorten(p APIShortenProcessor, l *zap.Logger, ep AuditEventPublisher) http.HandlerFunc {
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

		resBody, userUUID, err := p.Process(r.Context(), req)
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

		ep.Publish(model.AuditEvent{
			TS:      time.Now().Unix(),
			Action:  model.AuditActionShorten,
			UserID:  userUUID,
			OrigURL: req.OrigURL,
		})
	}
}
