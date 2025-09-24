package handler

import (
	"fmt"
)

func validateContentType(ct string, allowed string) error {
	if ct == allowed {
		return nil
	}
	return fmt.Errorf("`Content-Type` is not allowed. requested: %s, allowed: %s", ct, allowed)
}
