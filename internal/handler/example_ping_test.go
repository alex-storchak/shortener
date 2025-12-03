package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/handler"
)

// mockPingProcessor is a stub for PingProcessor.
type mockPingProcessor struct {
	err error
}

func (m *mockPingProcessor) Process() error {
	return m.err
}

// This example demonstrates a successful response from the handler created by
// HandlePing. The handler returns status 200 (OK) when the service is healthy.
func ExampleHandlePing_ok() {
	mp := &mockPingProcessor{}
	logger := zap.NewNop()

	h := handler.HandlePing(mp, logger)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 200
}

// This example demonstrates an internal error response from the handler created
// by HandlePing when the health check fails. The handler returns
// status 500 (Internal Server Error).
func ExampleHandlePing_internalServerError() {
	mp := &mockPingProcessor{
		err: fmt.Errorf("db not available"),
	}
	logger := zap.NewNop()

	h := handler.HandlePing(mp, logger)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 500
}
