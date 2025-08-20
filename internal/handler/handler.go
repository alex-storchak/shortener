package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/handler/config"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func Serve(cfg *config.Config, shortener service.IShortener) error {
	h := newHandlers(shortener, cfg.BaseURL)
	router := newRouter(h)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	return srv.ListenAndServe()
}

func newRouter(h *handlers) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", h.MainPageHandler)
	r.Get("/{id:[a-zA-Z0-9_-]+}", h.ShortURLHandler)

	return r
}

type handlers struct {
	shortener service.IShortener
	baseURL   string
}

func newHandlers(shortener service.IShortener, baseURL string) *handlers {
	return &handlers{
		shortener: shortener,
		baseURL:   baseURL,
	}
}

func (h *handlers) MainPageHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	shortID, err := h.shortener.Shorten(string(body))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	shortURL := fmt.Sprintf("%s/%s", h.baseURL, shortID)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))
}

func (h *handlers) ShortURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	shortID := chi.URLParam(req, "id")
	targetURL, err := h.shortener.Extract(shortID)
	if err != nil {
		if errors.Is(err, service.ErrShortenerShortIDNotFound) {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", targetURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
