package server

import (
	"context"

	"teiwrappergolang/pkg/client"
	pb "teiwrappergolang/protos/gen/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the TextEmbeddingsService gRPC service
type Server struct {
	pb.UnimplementedTextEmbeddingsServiceServer
	client *client.Client
	logger *zap.Logger
}

// NewServer creates a new gRPC server
func NewServer(client *client.Client, logger *zap.Logger) *Server {
	return &Server{
		client: client,
		logger: logger.Named("grpc-server"),
	}
}

// Embed implements the Embed RPC
func (s *Server) Embed(ctx context.Context, req *pb.EmbedRequest) (*pb.EmbedResponse, error) {
	s.logger.Debug("Embed RPC called", zap.Int("inputs_count", len(req.Inputs)))

	// Convert protobuf request to domain request
	domainReq, err := s.convertEmbedRequest(req)
	if err != nil {
		s.logger.Error("Failed to convert embed request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	// Call domain service
	domainResp, err := s.client.Embed(ctx, domainReq)
	if err != nil {
		s.logger.Error("Embed operation failed", zap.Error(err))
		return nil, s.convertError(err)
	}

	// Convert domain response to protobuf response
	pbResp := s.convertEmbedResponse(domainResp)

	s.logger.Debug("Embed RPC completed", zap.Int("embeddings_count", len(pbResp.Embeddings)))
	return pbResp, nil
}

// EmbedAll implements the EmbedAll RPC
func (s *Server) EmbedAll(ctx context.Context, req *pb.EmbedAllRequest) (*pb.EmbedAllResponse, error) {
	s.logger.Debug("EmbedAll RPC called", zap.Int("inputs_count", len(req.Inputs)))

	domainReq, err := s.convertEmbedAllRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	domainResp, err := s.client.EmbedAll(ctx, domainReq)
	if err != nil {
		s.logger.Error("EmbedAll operation failed", zap.Error(err))
		return nil, s.convertError(err)
	}

	pbResp := s.convertEmbedAllResponse(domainResp)
	return pbResp, nil
}

// EmbedSparse implements the EmbedSparse RPC
func (s *Server) EmbedSparse(ctx context.Context, req *pb.EmbedSparseRequest) (*pb.EmbedSparseResponse, error) {
	s.logger.Debug("EmbedSparse RPC called", zap.Int("inputs_count", len(req.Inputs)))

	domainReq, err := s.convertEmbedSparseRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	domainResp, err := s.client.EmbedSparse(ctx, domainReq)
	if err != nil {
		s.logger.Error("EmbedSparse operation failed", zap.Error(err))
		return nil, s.convertError(err)
	}

	pbResp := s.convertEmbedSparseResponse(domainResp)
	return pbResp, nil
}

// CalculateSimilarity implements the CalculateSimilarity RPC
func (s *Server) CalculateSimilarity(ctx context.Context, req *pb.SimilarityRequest) (*pb.SimilarityResponse, error) {
	s.logger.Debug("CalculateSimilarity RPC called",
		zap.String("source", req.SourceSentence[:min(50, len(req.SourceSentence))]),
		zap.Int("sentences_count", len(req.Sentences)),
	)

	domainReq, err := s.convertSimilarityRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	domainResp, err := s.client.CalculateSimilarity(ctx, domainReq)
	if err != nil {
		s.logger.Error("CalculateSimilarity operation failed", zap.Error(err))
		return nil, s.convertError(err)
	}

	pbResp := &pb.SimilarityResponse{
		Similarities: domainResp.Similarities,
	}

	return pbResp, nil
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
