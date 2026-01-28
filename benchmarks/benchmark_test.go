package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/teris-io/shortid"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/audit"
	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/handler/processor"
	"github.com/alex-storchak/shortener/internal/logger"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/random"
	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/alex-storchak/shortener/internal/service"
)

const (
	testURLsCount = 100000
)

var (
	testURLs []string
)

func init() {
	testURLs = generateTestURLs(testURLsCount)
}

func createTestApp() http.Handler {
	cfg := &config.Config{
		Logger: config.Logger{LogLevel: "warn"},
		Auth: config.Auth{
			CookieName:  config.DefAuthCookieName,
			TokenMaxAge: config.DefAuthTokenMaxAge,
		},
	}
	zl, err := logger.New(&cfg.Logger)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	storage := repository.NewMemoryURLStorage(zl)
	generator, err := shortid.New(1, shortid.DefaultABC, 1)
	if err != nil {
		zl.Error("failed to init generator", zap.Error(err))
	}
	shortener := service.NewShortener(generator, storage, zl)

	var observers []audit.Observer
	auditPublisher := audit.NewEventManager(observers, cfg.Audit, zl)
	ub := service.NewURLBuilder(cfg.Handler.BaseURL)
	shortenProc := processor.NewShorten(shortener, zl, ub, auditPublisher)
	expandProc := processor.NewExpand(shortener, zl, auditPublisher)
	pingProc := processor.NewPing(shortener, zl)
	apiShortenProc := processor.NewAPIShorten(shortener, zl, ub, auditPublisher)
	apiShortenBatchProc := processor.NewAPIShortenBatch(shortener, zl, ub)
	apiUserURLsProc := processor.NewAPIUserURLs(shortener, zl, ub)

	userStorage := repository.NewMemoryUserStorage(zl)
	authService := service.NewAuthService(zl, userStorage, &cfg.Auth)
	userManager := repository.NewUserManager(zl, userStorage)
	authResolver := service.NewAuthUserResolver(authService, userManager, &cfg.Auth)
	grpcUserResolver := service.NewAuthUserResolver(authService, userManager, &cfg.Auth)

	hDeps := &handler.ServerDeps{
		Logger:              zl,
		Config:              cfg,
		HTTPUserResolver:    authResolver,
		GRPCUserResolver:    grpcUserResolver,
		ShortenProc:         shortenProc,
		ExpandProc:          expandProc,
		PingProc:            pingProc,
		APIShortenProc:      apiShortenProc,
		APIShortenBatchProc: apiShortenBatchProc,
		APIUserURLsProc:     apiUserURLsProc,
	}

	return handler.NewRouter(hDeps)
}

func BenchmarkShorten(b *testing.B) {
	handlerToTest := createTestApp()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		b.StopTimer()
		idx := rand.IntN(testURLsCount)
		url := testURLs[idx]

		req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(url))
		req.Header.Set("Content-Type", "text/plain")

		rr := httptest.NewRecorder()
		b.StartTimer()

		handlerToTest.ServeHTTP(rr, req)
	}
}

func BenchmarkAPIShorten(b *testing.B) {
	handlerToTest := createTestApp()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		b.StopTimer()
		idx := rand.IntN(testURLsCount)
		url := testURLs[idx]
		reqData := model.ShortenRequest{OrigURL: url}
		jsonData, _ := json.Marshal(reqData)

		req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		b.StartTimer()

		handlerToTest.ServeHTTP(rr, req)
	}
}

func BenchmarkExpand(b *testing.B) {
	handlerToTest := createTestApp()

	shortURLs, err := seedShortenURLs(handlerToTest, nil, 10000)
	if err != nil {
		b.Fatal("failed to seed shorten URLs", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		b.StopTimer()
		idx := rand.IntN(len(shortURLs))
		shortID := shortURLs[idx]
		if shortID == "" {
			b.StartTimer()
			continue
		}

		req, _ := http.NewRequest("GET", "/"+shortID, nil)
		rr := httptest.NewRecorder()

		routeContext := chi.NewRouteContext()
		routeContext.URLParams.Add(handler.ShortIDParam, shortID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext))
		b.StartTimer()

		handlerToTest.ServeHTTP(rr, req)
	}
}

func BenchmarkApiUserURLsGet(b *testing.B) {
	handlerToTest := createTestApp()

	userCookie := getUserCookie(handlerToTest)
	if userCookie == nil {
		b.Fatal("Auth middleware did not set the expected 'auth' cookie")
	}

	if _, err := seedShortenURLs(handlerToTest, userCookie, 50); err != nil {
		b.Fatal("failed to seed user shorten URLs", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		b.StopTimer()
		req, _ := http.NewRequest("GET", "/api/user/urls", nil)
		req.AddCookie(userCookie)

		rr := httptest.NewRecorder()
		b.StartTimer()

		handlerToTest.ServeHTTP(rr, req)
	}
}

func BenchmarkAPIShortenBatch(b *testing.B) {
	handlerToTest := createTestApp()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		b.StopTimer()
		batchItems := make([]model.BatchShortenRequestItem, 10)
		for j := 0; j < 10; j++ {
			idx := rand.IntN(testURLsCount)
			batchItems[j] = model.BatchShortenRequestItem{
				CorrelationID: fmt.Sprintf("corr-%d", j),
				OriginalURL:   testURLs[idx],
			}
		}

		jsonData, _ := json.Marshal(batchItems)

		req, _ := http.NewRequest("POST", "/api/shorten/batch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		b.StartTimer()

		handlerToTest.ServeHTTP(rr, req)
	}
}

func BenchmarkApiUserURLsDelete(b *testing.B) {
	handlerToTest := createTestApp()

	userCookie := getUserCookie(handlerToTest)
	if userCookie == nil {
		b.Fatal("Auth middleware did not set the expected 'auth' cookie")
	}

	shortIDsToDelete, err := seedShortenURLs(handlerToTest, userCookie, 1000)
	if err != nil {
		b.Fatal("failed to seed user shorten URLs", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		b.StopTimer()

		batchSize := 10
		idsBatch := make([]string, 0, batchSize)
		for j := 0; j < batchSize; j++ {
			idx := rand.IntN(len(shortIDsToDelete))
			idsBatch = append(idsBatch, shortIDsToDelete[idx])
		}

		jsonData, _ := json.Marshal(idsBatch)

		req, _ := http.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(userCookie)

		rr := httptest.NewRecorder()
		b.StartTimer()

		handlerToTest.ServeHTTP(rr, req)
	}
}

func getUserCookie(handlerToTest http.Handler) *http.Cookie {
	for i := 0; i < len(testURLs); i++ {
		url := testURLs[i]
		reqData := model.ShortenRequest{OrigURL: url}
		jsonData, _ := json.Marshal(reqData)

		req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
			continue
		}

		resp := rr.Result()
		for _, cookie := range resp.Cookies() {
			if cookie.Name == config.DefAuthCookieName {
				if err := resp.Body.Close(); err != nil {
					return nil
				}
				return cookie
			}
		}
	}
	return nil
}

func seedShortenURLs(handlerToTest http.Handler, userCookie *http.Cookie, count int) ([]string, error) {
	shortIDs := make([]string, 0)

	for i := 0; i < count; i++ {
		url := testURLs[i]
		reqData := model.ShortenRequest{OrigURL: url}
		jsonData, _ := json.Marshal(reqData)

		req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		if userCookie != nil {
			req.AddCookie(userCookie)
		}

		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			continue
		}

		var result model.ShortenResponse
		if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response body: %w", err)
		}
		shortIDs = append(shortIDs, result.ShortURL)
	}
	if len(shortIDs) == 0 {
		return nil, fmt.Errorf("no short IDs were generated for user")
	}
	return shortIDs, nil
}

func generateTestURLs(count int) []string {
	urls := make([]string, count)
	for i := 0; i < count; i++ {
		urls[i] = random.URL().String()
	}
	return urls
}
