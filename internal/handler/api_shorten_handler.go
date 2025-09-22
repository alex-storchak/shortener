package handler

import (
	"errors"
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
	if errors.Is(err, service.ErrEmptyURL) {
		res.WriteHeader(http.StatusBadRequest)
		return
	} else if errors.Is(err, service.ErrURLAlreadyExists) {
		err = h.writeResponse(res, http.StatusConflict, resBody)
		if err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}
		return
	} else if errors.Is(err, service.ErrJSONDecode) {
		h.logger.Error("failed to shorten because of failed to decode request json", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	} else if err != nil {
		h.logger.Error("failed to shorten, unknown error", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.writeResponse(res, http.StatusCreated, resBody)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

func (h *APIShortenHandler) writeResponse(res http.ResponseWriter, status int, resBody *model.ShortenResponse) error {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	if err := h.jsonEncoder.Encode(res, resBody); err != nil {
		return err
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
