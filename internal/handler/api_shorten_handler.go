package handler

import (
	"errors"
	"net/http"

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
	logger = logger.With(
		zap.String("component", "handler"),
		zap.String("handler", "api_shorten"),
	)
	return &APIShortenHandler{
		apiShortenSrv: apiShortenService,
		jsonEncoder:   jsonEncoder,
		logger:        logger,
	}
}

func (h *APIShortenHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if err := validateMethod(req.Method, http.MethodPost); err != nil {
		h.logger.Error("Validation of request method failed", zap.Error(err))
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ct := req.Header.Get("Content-Type")
	if err := validateContentType(ct, "application/json"); err != nil {
		h.logger.Error("Validation of `Content-Type` failed", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	resBody, err := h.apiShortenSrv.Shorten(req.Body)
	if errors.Is(err, service.ErrEmptyURL) {
		h.logger.Error("failed to shorten because of empty url in request json", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
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

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	h.logger.Debug("sending HTTP 201 response")

	if err := h.jsonEncoder.Encode(res, resBody); err != nil {
		h.logger.Debug("error encoding response", zap.Error(err))
		return
	}
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
