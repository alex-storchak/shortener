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

// mockAPIShortenBatchProcessor — stub for APIShortenBatchProcessor.
type mockAPIShortenBatchProcessor struct {
	resp model.BatchShortenResponse
	err  error
}

func (m *mockAPIShortenBatchProcessor) Process(
	_ context.Context,
	_ model.BatchShortenRequest,
) (model.BatchShortenResponse, error) {
	return m.resp, m.err
}

// This example demonstrates the successful response from a handler
// created by the HandleAPIShortenBatch function.
// Handler returns response status 201 (Created) and JSON body with short URLs.
func ExampleHandleAPIShortenBatch_created() {
	body := []byte(`[
		{"correlation_id": "1", "original_url": "https://example.com/one"},
		{"correlation_id": "2", "original_url": "https://example.com/two"}
	]`)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	mp := &mockAPIShortenBatchProcessor{
		resp: model.BatchShortenResponse{
			{
				CorrelationID: "1",
				ShortURL:      "http://short/aaa",
			},
			{
				CorrelationID: "2",
				ShortURL:      "http://short/bbb",
			},
		},
	}

	logger := zap.NewNop()
	h := handler.HandleAPIShortenBatch(mp, logger)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Output:
	// 201
	// [{"correlation_id":"1","short_url":"http://short/aaa"},{"correlation_id":"2","short_url":"http://short/bbb"}]
}

// This example demonstrates the error response from a handler
// created by the HandleAPIShortenBatch function because of invalid `Content-Type`.
// Handler returns response status 400 (Bad Request) and empty body.
func ExampleHandleAPIShortenBatch_badRequestOnInvalidContentType() {
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", nil)
	// Content-Type has not set thus validateContentType will fail

	mp := &mockAPIShortenBatchProcessor{}
	logger := zap.NewNop()
	h := handler.HandleAPIShortenBatch(mp, logger)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 400
	// true
}

// This example demonstrates the error response from a handler
// created by the HandleAPIShortenBatch function because of empty input batch.
// Handler returns response status 400 (Bad Request) and empty body.
func ExampleHandleAPIShortenBatch_badRequestOnEmptyInputError() {
	body := []byte(`[]`)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	mp := &mockAPIShortenBatchProcessor{
		err: service.ErrEmptyInputBatch,
	}
	logger := zap.NewNop()
	h := handler.HandleAPIShortenBatch(mp, logger)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 400
	// true
}

// This example demonstrates the error response from a handler
// created by the HandleAPIShortenBatch function because of internal error.
// Handler returns response status 500 (Internal Server Error) and empty body.
func ExampleHandleAPIShortenBatch_internalServerError() {
	body := []byte(`[
		{"correlation_id": "1", "original_url": "https://example.com/one"}
	]`)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	mp := &mockAPIShortenBatchProcessor{
		err: fmt.Errorf("storage error"),
	}
	logger := zap.NewNop()
	h := handler.HandleAPIShortenBatch(mp, logger)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 500
	// true
}
