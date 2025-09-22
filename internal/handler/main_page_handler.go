package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type MainPageHandler struct {
	mainPageSrv service.IMainPageService
	logger      *zap.Logger
}

func NewMainPageHandler(mainPageService service.IMainPageService, logger *zap.Logger) *MainPageHandler {
	return &MainPageHandler{
		mainPageSrv: mainPageService,
		logger:      logger,
	}
}

func (h *MainPageHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.mainPageSrv.Shorten(body)
	if errors.Is(err, service.ErrEmptyBody) {
		res.WriteHeader(http.StatusBadRequest)
		return
	} else if errors.Is(err, service.ErrURLAlreadyExists) {
		h.writeResponse(res, http.StatusConflict, shortURL)
		return
	} else if err != nil {
		h.logger.Error("failed to process main page request", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.writeResponse(res, http.StatusCreated, shortURL)
}

func (h *MainPageHandler) writeResponse(res http.ResponseWriter, status int, shortURL string) {
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(status)
	_, _ = res.Write([]byte(shortURL))
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
