package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type MainPageProcessor interface {
	Process(ctx context.Context, body []byte) (shortURL string, err error)
}

type MainPageHandler struct {
	srv    MainPageProcessor
	logger *zap.Logger
}

func NewMainPageHandler(srv MainPageProcessor, l *zap.Logger) *MainPageHandler {
	return &MainPageHandler{
		srv:    srv,
		logger: l,
	}
}

func (h *MainPageHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.srv.Process(req.Context(), body)
	if errors.Is(err, service.ErrEmptyInputURL) {
		res.WriteHeader(http.StatusBadRequest)
		return
	} else if errors.Is(err, service.ErrURLAlreadyExists) {
		if err := h.writeResponse(res, http.StatusConflict, shortURL); err != nil {
			h.logger.Error("failed to write response (status conflict) for main page request", zap.Error(err))
		}
		return
	} else if err != nil {
		h.logger.Error("failed to process main page request", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.writeResponse(res, http.StatusCreated, shortURL); err != nil {
		h.logger.Error("failed to write response (status created) for main page request", zap.Error(err))
	}
}

func (h *MainPageHandler) writeResponse(res http.ResponseWriter, status int, shortURL string) error {
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(status)
	if _, err := res.Write([]byte(shortURL)); err != nil {
		return fmt.Errorf("write response `%s`: %w", shortURL, err)
	}
	return nil
}

func (h *MainPageHandler) Routes() []Route {
	return []Route{
		{
			Method:  http.MethodPost,
			Pattern: "/",
			Handler: h.ServeHTTP,
		},
	}
}
