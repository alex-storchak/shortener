package handler

import (
	"net/http"

	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type APIUserURLsHandler struct {
	srv     service.IAPIUserURLsService
	jsonEnc service.IJSONEncoder
	logger  *zap.Logger
}

func NewAPIUserURLsHandler(
	srv service.IAPIUserURLsService,
	jsonEncoder service.IJSONEncoder,
	logger *zap.Logger,
) *APIUserURLsHandler {
	return &APIUserURLsHandler{
		srv:     srv,
		jsonEnc: jsonEncoder,
		logger:  logger,
	}
}

func (h *APIUserURLsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	respItems, err := h.srv.GetUserURLs(req.Context())
	if err != nil {
		h.logger.Error("error getting user urls", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	} else if len(*respItems) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	if err := h.jsonEnc.Encode(res, respItems); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		return
	}
}

func (h *APIUserURLsHandler) Routes() []Route {
	return []Route{
		{
			Method:  http.MethodGet,
			Pattern: "/api/user/urls",
			Handler: h.ServeHTTP,
		},
	}
}
