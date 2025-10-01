package handler

import (
	"errors"
	"net/http"

	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type APIShortenBatchHandler struct {
	batchSrv service.IAPIShortenBatchService
	jsonEnc  service.IJSONEncoder
	logger   *zap.Logger
}

func NewAPIShortenBatchHandler(
	batchService service.IAPIShortenBatchService,
	jsonEncoder service.IJSONEncoder,
	logger *zap.Logger,
) *APIShortenBatchHandler {
	return &APIShortenBatchHandler{
		batchSrv: batchService,
		jsonEnc:  jsonEncoder,
		logger:   logger,
	}
}

func (h *APIShortenBatchHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ct := req.Header.Get("Content-Type")
	if err := validateContentType(ct, "application/json"); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	respItems, err := h.batchSrv.ShortenBatch(req.Context(), req.Body)
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
