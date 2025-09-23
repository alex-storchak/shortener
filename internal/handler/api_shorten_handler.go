package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type APIShortenHandler struct {
	apiShortenSrv service.IAPIShortenService
	jsonEncoder   service.IJSONEncoder
	logger        *zap.Logger
}

func NewAPIShortenHandler(
	apiShortenService service.IAPIShortenService,
	jsonEncoder service.IJSONEncoder,
	logger *zap.Logger,
) *APIShortenHandler {
	return &APIShortenHandler{
		apiShortenSrv: apiShortenService,
		jsonEncoder:   jsonEncoder,
		logger:        logger,
	}
}

func (h *APIShortenHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ct := req.Header.Get("Content-Type")
	if err := validateContentType(ct, "application/json"); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	resBody, err := h.apiShortenSrv.Shorten(req.Body)
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
	if err := h.jsonEncoder.Encode(res, resBody); err != nil {
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
