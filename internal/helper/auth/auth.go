// Package auth provides authentication and authorization utilities
// for handling user context in HTTP requests.
package auth

import (
	"context"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
)

type userCtxKey struct{}

// WithUser adds a model.User to the context for authentication purposes.
// This is typically used in middleware to attach user information
// to incoming requests.
func WithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

// GetCtxUserUUID retrieves the user UUID from the context.
// Returns an error if no user is found in the context or if the
// stored value has an unexpected type.
func GetCtxUserUUID(ctx context.Context) (string, error) {
	value := ctx.Value(userCtxKey{})
	if value == nil {
		return "", fmt.Errorf("user not found in context")
	}

	if user, ok := value.(*model.User); ok {
		return user.UUID, nil
	}
	return "", fmt.Errorf("unexpected user type in context: %T", value)
}
