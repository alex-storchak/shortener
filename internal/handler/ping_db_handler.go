package handler

import (
	"errors"
	"net/http"

	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type PingDBHandler struct {
	pingDBSrv service.IPingDBService
	logger    *zap.Logger
}

func NewPingDBHandler(s service.IPingDBService, l *zap.Logger) *PingDBHandler {
	l = l.With(
		zap.String("component", "handler"),
		zap.String("handler", "ping_db"),
	)
	return &PingDBHandler{
		pingDBSrv: s,
		logger:    l,
	}
}

func (h *PingDBHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if err := validateMethod(req.Method, http.MethodGet); err != nil {
		h.logger.Error("Validation of request method failed", zap.Error(err))
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := h.pingDBSrv.Ping()
	if errors.Is(err, service.ErrFailedToPingDB) {
		h.logger.Error("failed to ping database", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	} else if err != nil {
		h.logger.Error("failed to ping database. unknown error", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (h *PingDBHandler) Routes() []Route {
	return []Route{
		{
			Method:  http.MethodGet,
			Pattern: "/ping",
			Handler: h.ServeHTTP,
		},
	}
}
