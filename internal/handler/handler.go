package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alex-storchak/shortener/internal/handler/config"
	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/gorilla/mux"
)

func Serve(cfg config.Config, shortener *service.Shortener) error {
	h := newHandlers(shortener)
	router := newRouter(h)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort),
		Handler: router,
	}

	return srv.ListenAndServe()
}

func newRouter(h *handlers) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", h.MainPageHandler)
	r.HandleFunc("/{id:[a-zA-Z0-9_-]+}", h.ShortURLHandler).Name("shortUrl")
	r.NotFoundHandler = http.HandlerFunc(h.NotFoundHandler)

	return r
}

type handlers struct {
	shortener *service.Shortener
}

func newHandlers(shortener *service.Shortener) *handlers {
	return &handlers{
		shortener: shortener,
	}
}

func (h *handlers) NotFoundHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNotFound)
}

func (h *handlers) MainPageHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	shortID, err := h.shortener.Shorten(string(body))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	shortURL := fmt.Sprintf("http://%s/%s", req.Host, shortID)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))
}

func (h *handlers) ShortURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(req)
	shortID, ok := vars["id"]
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	targetURL, err := h.shortener.Extract(shortID)
	if err != nil {
		if errors.Is(err, repository.ErrShortURLNotFound) {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", targetURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
