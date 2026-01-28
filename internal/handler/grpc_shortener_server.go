package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "github.com/alex-storchak/shortener/api/proto/shortener"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/alex-storchak/shortener/internal/service"
)

type GRPCShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	logger       *zap.Logger
	shortenProc  APIShortenProcessor
	expandProc   ExpandProcessor
	userURLsProc APIUserURLsProcessor
}

func NewGRPCShortenerServer(deps *ServerDeps) *GRPCShortenerServer {
	server := GRPCShortenerServer{
		logger:       deps.Logger,
		shortenProc:  deps.APIShortenProc,
		expandProc:   deps.ExpandProc,
		userURLsProc: deps.APIUserURLsProc,
	}
	return &server
}

func (s *GRPCShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	r := model.ShortenRequest{
		OrigURL: req.GetUrl(),
	}
	result, err := s.shortenProc.Process(ctx, r)
	if errors.Is(err, service.ErrEmptyInputURL) {
		return nil, status.Error(codes.InvalidArgument, "empty input url")
	} else if errors.Is(err, service.ErrURLAlreadyExists) {
		return nil, status.Error(codes.AlreadyExists, "url already exists")
	} else if err != nil {
		s.logger.Error("failed to shorten", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	res := pb.URLShortenResponse_builder{
		Result: proto.String(result.ShortURL),
	}.Build()

	return res, nil
}

func (s *GRPCShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	shortID := req.GetId()

	origURL, err := s.expandProc.Process(ctx, shortID)
	var nfErr *repository.DataNotFoundError
	if errors.As(err, &nfErr) {
		return nil, status.Error(codes.NotFound, "url not found")
	} else if errors.Is(err, repository.ErrDataDeleted) {
		return nil, status.Error(codes.FailedPrecondition, "data is already deleted")
	} else if err != nil {
		s.logger.Error("failed to expand short url", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	res := pb.URLExpandResponse_builder{
		Result: proto.String(origURL),
	}.Build()
	return res, nil
}

func (s *GRPCShortenerServer) ListUserURLs(ctx context.Context, _ *pb.UserURLsRequest) (*pb.UserURLsResponse, error) {
	respItems, err := s.userURLsProc.ProcessGet(ctx)
	if err != nil {
		s.logger.Error("error getting user urls", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	urlDataList := make([]*pb.URLData, 0, len(respItems))
	for _, item := range respItems {
		data := pb.URLData_builder{
			ShortUrl:    proto.String(item.ShortURL),
			OriginalUrl: proto.String(item.OrigURL),
		}.Build()
		urlDataList = append(urlDataList, data)
	}

	res := pb.UserURLsResponse_builder{
		Url: urlDataList,
	}.Build()

	return res, nil
}
