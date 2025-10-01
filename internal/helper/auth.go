package helper

import (
	"context"
	"fmt"

	"github.com/alex-storchak/shortener/internal/model"
)

type UserCtxKey struct{}

func GetCtxUserUUID(ctx context.Context) (string, error) {
	value := ctx.Value(UserCtxKey{})
	if value == nil {
		return "", fmt.Errorf("user not found in context")
	}

	if user, ok := value.(*model.User); ok {
		return user.UUID, nil
	}
	return "", fmt.Errorf("unexpected user type in context: %T", value)
}
