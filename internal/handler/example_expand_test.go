package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/repository"
)

// mockExpandProcessor is a stub for ExpandProcessor.
type mockExpandProcessor struct {
	origURL string
	userID  string
	err     error
}

func (m *mockExpandProcessor) Process(_ context.Context, _ string) (string, string, error) {
	return m.origURL, m.userID, m.err
}

// This example demonstrates a successful response from the handler created by
// HandleExpand. The handler returns status 307 (Temporary Redirect) and
// sets the Location header to the original URL.
func ExampleHandleExpand_temporaryRedirect() {
	mp := &mockExpandProcessor{
		origURL: "https://example.com/one",
		userID:  "user-123",
	}
	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}

	h := handler.HandleExpand(mp, logger, ap)

	shortID := "aaa"
	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	rr := httptest.NewRecorder()

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(handler.ShortIDParam, shortID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext))

	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Location"))

	// Output:
	// 307
	// https://example.com/one
}

// This example demonstrates a 404 Not Found response from the handler created by
// HandleExpand when the short ID does not exist.
func ExampleHandleExpand_notFound() {
	mp := &mockExpandProcessor{
		err: &repository.DataNotFoundError{},
	}
	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}

	h := handler.HandleExpand(mp, logger, ap)

	shortID := "unknown"
	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	rr := httptest.NewRecorder()

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(handler.ShortIDParam, shortID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext))

	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 404
}

// This example demonstrates a 410 Gone response from the handler created by
// HandleExpand when the URL has been deleted.
func ExampleHandleExpand_gone() {
	mp := &mockExpandProcessor{
		err: repository.ErrDataDeleted,
	}
	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}

	h := handler.HandleExpand(mp, logger, ap)

	shortID := "deleted"
	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	rr := httptest.NewRecorder()

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(handler.ShortIDParam, shortID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext))

	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 410
}

// This example demonstrates a 500 (Internal Server Error) response from the handler created by
// HandleExpand when a storage or service error occurs.
func ExampleHandleExpand_internalServerError() {
	mp := &mockExpandProcessor{
		err: errors.New("storage error"),
	}
	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}

	h := handler.HandleExpand(mp, logger, ap)

	shortID := "err"
	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	rr := httptest.NewRecorder()

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(handler.ShortIDParam, shortID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext))
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 500
}
