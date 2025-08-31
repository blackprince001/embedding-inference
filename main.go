package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"teiwrappergolang/internal/config"
	"teiwrappergolang/internal/infrastructure/logging"
	"teiwrappergolang/internal/infrastructure/wrapper"
	"teiwrappergolang/internal/server"

	pb "teiwrappergolang/protos/gen/v1"

	"teiwrappergolang/pkg/client"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	logCfg := cfg.Log
	logger, err := logging.NewLogger(&logCfg)
	if err != nil {
		log.Fatalf("failed to create logger: %s", err)
	}

	teiCfg := cfg.TEI
	httpClient, err := wrapper.NewHTTPClient(&teiCfg, logger)
	if err != nil {
		log.Fatalf("failed to create HTTP client: %s", err)
	}

	client := client.NewClient(cfg, httpClient, logger)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger.Logger)),
		grpc.MaxRecvMsgSize(16*1024*1024), // 16MB max message size
		grpc.MaxSendMsgSize(16*1024*1024),
	)

	textEmbeddingsServer := server.NewServer(client, logger.Logger)
	pb.RegisterTextEmbeddingsServiceServer(grpcServer, textEmbeddingsServer)

	reflection.Register(grpcServer)

	ls, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	if err := grpcServer.Serve(ls); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}

func loggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		logger.Info("Received gRPC request",
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
		)

		resp, err = handler(ctx, req)

		if err != nil {
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC request succeeded",
				zap.String("method", info.FullMethod),
				zap.Any("response", resp),
			)
		}

		return resp, err
	}
}
