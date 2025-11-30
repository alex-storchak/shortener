package handler_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

// mockAPIShortenProcessor is a stub for APIShortenProcessor.
type mockAPIShortenProcessor struct {
	resp    *model.ShortenResponse
	userID  string
	err     error
	lastReq model.ShortenRequest
}

func (m *mockAPIShortenProcessor) Process(
	_ context.Context,
	req model.ShortenRequest,
) (*model.ShortenResponse, string, error) {
	m.lastReq = req
	return m.resp, m.userID, m.err
}

// mockAuditEventPublisher is a stub for AuditEventPublisher.
type mockAuditEventPublisher struct{}

func (m *mockAuditEventPublisher) Publish(_ model.AuditEvent) { /*no-op*/ }

// This example demonstrates a successful response from the handler created by
// HandleAPIShorten. The handler returns status 201 (Created)
// and a JSON body with the short URL.
func ExampleHandleAPIShorten_created() {
	body := []byte(`{"url":"https://example.com/one"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	mp := &mockAPIShortenProcessor{
		resp: &model.ShortenResponse{
			ShortURL: "http://short/aaa",
		},
		userID: "user-123",
	}

	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}
	h := handler.HandleAPIShorten(mp, logger, ap)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Output:
	// 201
	// {"result":"http://short/aaa"}
}

// This example demonstrates a bad request response from the handler created by
// HandleAPIShorten due to invalid Content-Type.
// The handler returns status 400 (Bad Request) and an empty body.
func ExampleHandleAPIShorten_badRequestOnInvalidContentType() {
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", nil)
	// Content-Type is not set, so validateContentType will fail.

	mp := &mockAPIShortenProcessor{}
	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}
	h := handler.HandleAPIShorten(mp, logger, ap)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 400
	// true
}

// This example demonstrates a bad request response from the handler created
// by HandleAPIShorten due to an empty URL.
// The handler returns status 400 (Bad Request) and an empty body.
func ExampleHandleAPIShorten_badRequestOnEmptyInputURL() {
	body := []byte(`{"url":""}`)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	mp := &mockAPIShortenProcessor{
		err: service.ErrEmptyInputURL,
	}

	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}
	h := handler.HandleAPIShorten(mp, logger, ap)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 400
	// true
}

// This example demonstrates a conflict response from the handler created
// by HandleAPIShorten when the URL already exists.
// The handler returns status 409 (Conflict) and a JSON body with
// the already existing short URL.
func ExampleHandleAPIShorten_conflictOnExistingURL() {
	body := []byte(`{"url":"https://example.com/existing"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	mp := &mockAPIShortenProcessor{
		resp: &model.ShortenResponse{
			ShortURL: "http://short/exist",
		},
		err:    service.ErrURLAlreadyExists,
		userID: "user-456",
	}

	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}
	h := handler.HandleAPIShorten(mp, logger, ap)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Output:
	// 409
	// {"result":"http://short/exist"}
}

// This example demonstrates an internal error response from the handler created
// by HandleAPIShorten due to a service/storage error.
// The handler returns status 500 (Internal Server Error) and an empty body.
func ExampleHandleAPIShorten_internalServerError() {
	body := []byte(`{"url":"https://example.com/one"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	mp := &mockAPIShortenProcessor{
		err: fmt.Errorf("storage error"),
	}

	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}
	h := handler.HandleAPIShorten(mp, logger, ap)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 500
	// true
}
