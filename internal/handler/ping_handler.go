package handler

import (
	"net/http"

	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type PingHandler struct {
	pingSrv service.IPingService
	logger  *zap.Logger
}

func NewPingHandler(s service.IPingService, l *zap.Logger) *PingHandler {
	return &PingHandler{
		pingSrv: s,
		logger:  l,
	}
}

func (h *PingHandler) ServeHTTP(res http.ResponseWriter, _ *http.Request) {
	err := h.pingSrv.Ping()
	if err != nil {
		h.logger.Error("failed to ping service", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (h *PingHandler) Routes() []Route {
	return []Route{
		{
			Method:  http.MethodGet,
			Pattern: "/ping",
			Handler: h.ServeHTTP,
		},
	}
}
