package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type ShortenProcessor interface {
	Process(ctx context.Context, r io.Reader) (*model.ShortenResponse, error)
}

type APIShortenHandler struct {
	srv    ShortenProcessor
	enc    service.Encoder
	logger *zap.Logger
}

func NewAPIShortenHandler(
	srv ShortenProcessor,
	enc service.Encoder,
	l *zap.Logger,
) *APIShortenHandler {
	return &APIShortenHandler{
		srv:    srv,
		enc:    enc,
		logger: l,
	}
}

func (h *APIShortenHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ct := req.Header.Get("Content-Type")
	if err := validateContentType(ct, "application/json"); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	resBody, err := h.srv.Process(req.Context(), req.Body)
	if errors.Is(err, service.ErrEmptyInputURL) {
		res.WriteHeader(http.StatusBadRequest)
		return
	} else if errors.Is(err, service.ErrURLAlreadyExists) {
		if err := h.writeResponse(res, http.StatusConflict, resBody); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}
		return
	} else if err != nil {
		h.logger.Error("failed to shorten", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.writeResponse(res, http.StatusCreated, resBody); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

func (h *APIShortenHandler) writeResponse(res http.ResponseWriter, status int, resBody *model.ShortenResponse) error {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	if err := h.enc.Encode(res, resBody); err != nil {
		return fmt.Errorf("failed to encode response body `%v`: %w", resBody, err)
	}
	return nil
}

func (h *APIShortenHandler) Routes() []Route {
	return []Route{
		{
			Method:  http.MethodPost,
			Pattern: "/api/shorten",
			Handler: h.ServeHTTP,
		},
	}
}
