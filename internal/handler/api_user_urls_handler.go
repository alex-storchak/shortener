package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type UserURLsProcessor interface {
	ProcessGet(ctx context.Context) ([]model.UserURLsResponseItem, error)
	ProcessDelete(ctx context.Context, r io.Reader) error
}

type APIUserURLsHandler struct {
	srv    UserURLsProcessor
	enc    service.Encoder
	logger *zap.Logger
}

func NewAPIUserURLsHandler(
	srv UserURLsProcessor,
	enc service.Encoder,
	l *zap.Logger,
) *APIUserURLsHandler {
	return &APIUserURLsHandler{
		srv:    srv,
		enc:    enc,
		logger: l,
	}
}

func (h *APIUserURLsHandler) ServeHTTPGet(res http.ResponseWriter, req *http.Request) {
	respItems, err := h.srv.ProcessGet(req.Context())
	if err != nil {
		h.logger.Error("error getting user urls", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	} else if len(respItems) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	if err := h.enc.Encode(res, respItems); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		return
	}
}

func (h *APIUserURLsHandler) ServeHTTPDelete(res http.ResponseWriter, req *http.Request) {
	if err := h.srv.ProcessDelete(req.Context(), req.Body); err != nil {
		h.logger.Error("error deleting user urls", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusAccepted)
}

func (h *APIUserURLsHandler) Routes() []Route {
	return []Route{
		{
			Method:  http.MethodGet,
			Pattern: "/api/user/urls",
			Handler: h.ServeHTTPGet,
		},
		{
			Method:  http.MethodDelete,
			Pattern: "/api/user/urls",
			Handler: h.ServeHTTPDelete,
		},
	}
}
