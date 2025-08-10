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
	r.HandleFunc("/{id:[a-zA-Z0-9_-]+}", h.ShortUrlHandler).Name("shortUrl")
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
	return
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

	shortId, err := h.shortener.Shorten(string(body))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	shortUrl := fmt.Sprintf("http://%s/%s", req.Host, shortId)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortUrl))
}

func (h *handlers) ShortUrlHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(req)
	shortId, ok := vars["id"]
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	targetUrl, err := h.shortener.Extract(shortId)
	if err != nil {
		if errors.Is(err, repository.ErrShortUrlNotFound) {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", targetUrl)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
