package auth

import (
	"context"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
)

type userCtxKey struct{}

func WithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

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
