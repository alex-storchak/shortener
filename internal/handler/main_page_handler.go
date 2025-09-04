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
	logger = logger.With(
		zap.String("component", "handler"),
		zap.String("handler", "main_page"),
	)
	return &MainPageHandler{
		mainPageSrv: mainPageService,
		logger:      logger,
	}
}

func (h *MainPageHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if err := validateMethod(req.Method, http.MethodPost); err != nil {
		h.logger.Error("Validation of request method failed", zap.Error(err))
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.mainPageSrv.Shorten(body)
	if errors.Is(err, service.ErrEmptyBody) {
		h.logger.Error("empty request body", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		h.logger.Error("failed to process main page request", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
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
