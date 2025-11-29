package handler

import (
	"fmt"
)

// validateContentType validates that the request Content-Type header matches the expected value.

// Parameters:
//   - ct: The actual Content-Type header value from the request
//   - allowed: The expected Content-Type value that is permitted
//
// Returns:
//   - error: Descriptive error if content type doesn't match, nil if validation passes
//
// Example:
//
//	err := validateContentType(r.Header.Get("Content-Type"), "application/json")
//	if err != nil {
//	    http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
//	    return
//	}
func validateContentType(ct string, allowed string) error {
	if ct == allowed {
		return nil
	}
	return fmt.Errorf("`Content-Type` is not allowed. requested: %s, allowed: %s", ct, allowed)
}
