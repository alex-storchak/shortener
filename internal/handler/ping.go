package handler

import (
	"net/http"

	"go.uber.org/zap"
)

type PingProcessor interface {
	Process() error
}

func handlePing(p PingProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		err := p.Process()
		if err != nil {
			l.Error("failed to ping service", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
