package handler

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/alex-storchak/shortener/api/proto/shortener"
	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/interceptor"
)

func ServeGRPC(deps *ServerDeps) (*grpc.Server, error) {
	cfg := deps.Config
	logger := deps.Logger

	server, err := newGrpcServer(cfg, logger, deps.GRPCUserResolver)
	if err != nil {
		return nil, fmt.Errorf("create grpc server: %w", err)
	}

	handler := NewGRPCShortenerServer(deps)
	pb.RegisterShortenerServiceServer(server, handler)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting server", zap.String("server address", cfg.Server.GRPCServerAddr))
		lis, err := net.Listen("tcp", cfg.Server.GRPCServerAddr)
		if err != nil {
			errCh <- fmt.Errorf("failed to listen: %w", err)
			return
		}
		if err := server.Serve(lis); err != nil {
			errCh <- fmt.Errorf("start to serve: %w", err)
		}
	}()

	// Проверяем, успешно ли запустился сервер
	select {
	case err := <-errCh:
		return nil, fmt.Errorf("start server: %w", err)
	case <-time.After(time.Second):
		return server, nil
	}
}

func newGrpcServer(cfg *config.Config, l *zap.Logger, ur interceptor.UserResolver) (*grpc.Server, error) {
	auth := interceptor.NewAuth(l, ur, cfg.Auth)
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(auth),
	}

	if cfg.Server.EnableHTTPS {
		if cfg.Server.SSLCertPath == "" {
			return nil, ErrEmptySSLCertPath
		}
		if cfg.Server.SSLKeyPath == "" {
			return nil, ErrEmptySSLKeyPath
		}
		creds, err := credentials.NewServerTLSFromFile(cfg.Server.SSLCertPath, cfg.Server.SSLKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	return grpc.NewServer(opts...), nil
}
