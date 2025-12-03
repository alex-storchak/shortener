// Package codec provides efficient JSON encoding and decoding utilities for HTTP handlers.
// It leverages the easyjson library for high-performance JSON serialization and deserialization
// with minimal memory allocations.
//
// The package is designed specifically for HTTP request/response handling and provides:
//   - Direct streaming of JSON responses to http.ResponseWriter
//   - Efficient reading and parsing of HTTP request bodies
//   - Proper content-type header management
//   - Comprehensive error handling with contextual error messages
//
// Usage:
//
//	// Encoding response
//	err := codec.EasyJSONEncode(w, http.StatusOK, &response)
//
//	// Decoding request
//	var request MyRequest
//	err := codec.EasyJSONDecode(r, &request)
package codec

import (
	"fmt"
	"io"
	"net/http"

	"github.com/mailru/easyjson"
)

// EasyJSONEncode encodes a type implementing easyjson.Marshaler directly to ResponseWriter.
// It sets the appropriate Content-Type header, writes the HTTP status code, and streams
// the JSON response efficiently without intermediate buffers.
//
// Parameters:
//   - w: HTTP response writer to write the JSON response
//   - status: HTTP status code to return (e.g., http.StatusOK, http.StatusCreated)
//   - v: Value implementing easyjson.Marshaler to be encoded as JSON
//
// Returns:
//   - error: nil on success, or error if encoding or writing fails
//
// Example:
//
//	resp := &MyResponse{Data: "example"}
//	if err := EasyJSONEncode(w, http.StatusOK, resp); err != nil {
//	    log.Printf("Failed to encode response: %v", err)
//	}
func EasyJSONEncode(w http.ResponseWriter, status int, v easyjson.Marshaler) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := easyjson.MarshalToWriter(v, w); err != nil {
		return fmt.Errorf("easyjson encode: %w", err)
	}
	return nil
}

// EasyJSONDecode reads the entire request body and unmarshals it into the provided value.
// It efficiently handles reading and closing the request body, then uses easyjson
// for high-performance JSON unmarshaling.
//
// Parameters:
//   - r: HTTP request containing the JSON body to decode
//   - v: Pointer to a value implementing easyjson.Unmarshaler to receive the decoded data
//
// Returns:
//   - error: nil on success, or error if reading, closing, or unmarshaling fails
//
// Note: The parameter v must be a pointer to a type that implements easyjson.Unmarshaler.
//
// Example:
//
//	var req MyRequest
//	if err := EasyJSONDecode(r, &req); err != nil {
//	    http.Error(w, "Invalid JSON", http.StatusBadRequest)
//	    return
//	}
func EasyJSONDecode(r *http.Request, v easyjson.Unmarshaler) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if cErr := r.Body.Close(); cErr != nil {
		return fmt.Errorf("close body: %w", cErr)
	}

	if err := easyjson.Unmarshal(data, v); err != nil {
		return fmt.Errorf("easyjson decode: %w", err)
	}
	return nil
}
