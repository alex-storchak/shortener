package handler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type ShortenBatchProcessor interface {
	Process(ctx context.Context, r io.Reader) ([]model.BatchShortenResponseItem, error)
}

type APIShortenBatchHandler struct {
	srv     ShortenBatchProcessor
	jsonEnc service.Encoder
	logger  *zap.Logger
}

func NewAPIShortenBatchHandler(
	srv ShortenBatchProcessor,
	enc service.Encoder,
	l *zap.Logger,
) *APIShortenBatchHandler {
	return &APIShortenBatchHandler{
		srv:     srv,
		jsonEnc: enc,
		logger:  l,
	}
}

func (h *APIShortenBatchHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ct := req.Header.Get("Content-Type")
	if err := validateContentType(ct, "application/json"); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	respItems, err := h.srv.Process(req.Context(), req.Body)
	if errors.Is(err, service.ErrEmptyInputURL) || errors.Is(err, service.ErrEmptyInputBatch) {
		res.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		h.logger.Error("failed to shorten batch", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	if err := h.jsonEnc.Encode(res, respItems); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		return
	}
}

func (h *APIShortenBatchHandler) Routes() []Route {
	return []Route{
		{
			Method:  http.MethodPost,
			Pattern: "/api/shorten/batch",
			Handler: h.ServeHTTP,
		},
	}
}
