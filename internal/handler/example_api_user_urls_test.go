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
)

// mockAPIUserURLsProcessor is a stub for APIUserURLsProcessor.
type mockAPIUserURLsProcessor struct {
	getResp model.UserURLsGetResponse
	getErr  error
	delErr  error
}

func (m *mockAPIUserURLsProcessor) ProcessGet(_ context.Context) (model.UserURLsGetResponse, error) {
	return m.getResp, m.getErr
}

func (m *mockAPIUserURLsProcessor) ProcessDelete(_ context.Context, _ model.UserURLsDelRequest) error {
	return m.delErr
}

// This example demonstrates a successful response from the handler created by
// HandleGetUserURLs when the user has URLs.
// The handler returns status 200 (OK) and a JSON body with the list of URLs.
func ExampleHandleGetUserURLs_ok() {
	mp := &mockAPIUserURLsProcessor{
		getResp: model.UserURLsGetResponse{
			{
				ShortURL: "http://short/aaa",
				OrigURL:  "https://example.com/one",
			},
			{
				ShortURL: "http://short/bbb",
				OrigURL:  "https://example.com/two",
			},
		},
	}

	logger := zap.NewNop()
	h := handler.HandleGetUserURLs(mp, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Content-Type"))
	// порядок полей у easyjson стабилен, поэтому можно проверить точное тело
	fmt.Println(rr.Body.String())

	// Output:
	// 200
	// application/json
	// [{"short_url":"http://short/aaa","original_url":"https://example.com/one"},{"short_url":"http://short/bbb","original_url":"https://example.com/two"}]
}

// This example demonstrates a 204 No Content response from the handler created by
// HandleGetUserURLs when the user has no URLs.
func ExampleHandleGetUserURLs_noContent() {
	mp := &mockAPIUserURLsProcessor{
		getResp: model.UserURLsGetResponse{},
	}

	logger := zap.NewNop()
	h := handler.HandleGetUserURLs(mp, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 204
	// true
}

// This example demonstrates a 500 (Internal Server Error) response from the handler created by
// HandleGetUserURLs when the processor returns an error.
func ExampleHandleGetUserURLs_internalServerError() {
	mp := &mockAPIUserURLsProcessor{
		getErr: fmt.Errorf("storage error"),
	}

	logger := zap.NewNop()
	h := handler.HandleGetUserURLs(mp, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String() == "")

	// Output:
	// 500
	// true
}

// This example demonstrates a successful request handling by the handler created by
// HandleDeleteUserURLs. The handler returns status 202 (Accepted) for a valid JSON body.
func ExampleHandleDeleteUserURLs_accepted() {
	body := []byte(`["aaa","bbb","ccc"]`)

	mp := &mockAPIUserURLsProcessor{}
	logger := zap.NewNop()
	h := handler.HandleDeleteUserURLs(mp, logger)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 202
}

// This example demonstrates a 400 Bad Request response from the handler created by
// HandleDeleteUserURLs when the request JSON is malformed.
func ExampleHandleDeleteUserURLs_badRequestOnDecodeError() {
	body := []byte(`["aaa", invalid, "ccc"]`) // invalid JSON

	mp := &mockAPIUserURLsProcessor{}
	logger := zap.NewNop()
	h := handler.HandleDeleteUserURLs(mp, logger)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 400
}

// This example demonstrates a 500 (Internal Server Error) response from the handler created by
// HandleDeleteUserURLs when the processor returns an error.
func ExampleHandleDeleteUserURLs_internalServerError() {
	body := []byte(`["aaa","bbb"]`)

	mp := &mockAPIUserURLsProcessor{
		delErr: fmt.Errorf("delete error"),
	}
	logger := zap.NewNop()
	h := handler.HandleDeleteUserURLs(mp, logger)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 500
}
