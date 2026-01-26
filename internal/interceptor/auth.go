package interceptor

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/helper/auth"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/service"
)

const (
	BearerPrefix = "Bearer "
)

type UserResolver interface {
	ResolveUserGRPC(tokenFromMetadata string) (user *model.User, err error)
}

func NewAuth(logger *zap.Logger, srv UserResolver, cfg config.Auth) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		var token string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get(cfg.CookieName)
			if len(values) > 0 {
				token = strings.TrimPrefix(values[0], BearerPrefix)
			}
		}

		user, err := srv.ResolveUserGRPC(token)
		if err != nil {
			if errors.Is(err, service.ErrUnauthorized) {
				return nil, status.Error(codes.Unauthenticated, "unauthorized")
			}
			logger.Error("failed to resolve auth user", zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}

		ctx = auth.WithUser(ctx, user)
		return handler(ctx, req)
	}
}
