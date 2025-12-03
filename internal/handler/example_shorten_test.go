package handler_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/service"
)

// mockShortenProcessor is a stub for ShortenProcessor.
type mockShortenProcessor struct {
	shortURL string
	userID   string
	err      error
}

func (m *mockShortenProcessor) Process(_ context.Context, _ []byte) (string, string, error) {
	return m.shortURL, m.userID, m.err
}

// This example demonstrates a successful response from the handler created by
// HandleShorten. The handler returns status 201 (Created) and a
// plain text body with the short URL.
func ExampleHandleShorten_created() {
	body := []byte("https://example.com/one")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	mp := &mockShortenProcessor{
		shortURL: "http://short/aaa",
		userID:   "user-123",
	}

	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}
	h := handler.HandleShorten(mp, logger, ap)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())
	fmt.Println(rr.Header().Get("Content-Type"))

	// Output:
	// 201
	// http://short/aaa
	// text/plain
}

// errReader is a helper type that always returns an error on Read.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("read error")
}

func (errReader) Close() error {
	return nil
}

// This example demonstrates a bad request response from the handler created
// by HandleShorten when the body read fails.
// The handler returns status 400 (Bad Request) and an empty body.
func ExampleHandleShorten_badRequestOnReadError() {
	// Use a Request with nil Body to force ReadAll error-like behavior.
	req := httptest.NewRequest(http.MethodPost, "/", errReader{})

	mp := &mockShortenProcessor{}
	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}
	h := handler.HandleShorten(mp, logger, ap)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 400
	// true
}

// This example demonstrates a bad request response from the handler created
// by HandleShorten due to an empty URL (ErrEmptyInputURL from processor).
// The handler returns status 400 (Bad Request) and an empty body.
func ExampleHandleShorten_badRequestOnEmptyInputURL() {
	body := []byte("")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	mp := &mockShortenProcessor{
		err: service.ErrEmptyInputURL,
	}

	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}
	h := handler.HandleShorten(mp, logger, ap)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 400
	// true
}

// This example demonstrates a conflict response from the handler created
// by HandleShorten when the URL already exists (ErrURLAlreadyExists).
// The handler returns status 409 (Conflict) and a plain text body with
// the already existing short URL.
func ExampleHandleShorten_conflictOnExistingURL() {
	body := []byte("https://example.com/existing")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	mp := &mockShortenProcessor{
		shortURL: "http://short/exist",
		err:      service.ErrURLAlreadyExists,
		userID:   "user-456",
	}

	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}
	h := handler.HandleShorten(mp, logger, ap)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())
	fmt.Println(rr.Header().Get("Content-Type"))

	// Output:
	// 409
	// http://short/exist
	// text/plain
}

// This example demonstrates an internal error response from the handler created
// by HandleShorten due to a service/storage error.
// The handler returns status 500 (Internal Server Error) and an empty body.
func ExampleHandleShorten_internalServerError() {
	body := []byte("https://example.com/one")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	mp := &mockShortenProcessor{
		err: fmt.Errorf("storage error"),
	}

	logger := zap.NewNop()
	ap := &mockAuditEventPublisher{}
	h := handler.HandleShorten(mp, logger, ap)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 500
	// true
}
