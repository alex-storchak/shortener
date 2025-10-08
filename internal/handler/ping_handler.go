package handler

import (
	"net/http"

	"go.uber.org/zap"
)

type PingProcessor interface {
	Process() error
}

type PingHandler struct {
	srv    PingProcessor
	logger *zap.Logger
}

func NewPingHandler(s PingProcessor, l *zap.Logger) *PingHandler {
	return &PingHandler{
		srv:    s,
		logger: l,
	}
}

func (h *PingHandler) ServeHTTP(res http.ResponseWriter, _ *http.Request) {
	err := h.srv.Process()
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
