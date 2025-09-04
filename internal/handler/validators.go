package handler

import (
	"fmt"
)

func validateMethod(method string, allowed string) error {
	if method == allowed {
		return nil
	}
	return fmt.Errorf("requested method is not allowed. requsted: %s, allowed: %s", method, allowed)
}

func validateContentType(ct string, allowed string) error {
	if ct == allowed {
		return nil
	}
	return fmt.Errorf("`Content-Type` is not allowed. requested: %s, allowed: %s", ct, allowed)
}
