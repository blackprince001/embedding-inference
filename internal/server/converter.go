package server

import (
	"github.com/blackprince001/embedding-inference/internal/domain/entities"
	"github.com/blackprince001/embedding-inference/internal/domain/errors"
	pb "github.com/blackprince001/embedding-inference/protos/gen/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) convertEmbedRequest(req *pb.EmbedRequest) (*entities.EmbedRequest, error) {
	domainReq := &entities.EmbedRequest{
		Inputs: entities.Input{Data: req.Inputs},
	}

	if req.Normalize != nil {
		domainReq.Normalize = req.Normalize
	}
	if req.PromptName != nil {
		domainReq.PromptName = req.PromptName
	}
	if req.Truncate != nil {
		domainReq.Truncate = req.Truncate
	}
	if req.TruncationDirection != nil {
		domainReq.TruncationDirection = convertTruncationDirection(*req.TruncationDirection)
	}

	return domainReq, nil
}

func (s *Server) convertEmbedAllRequest(req *pb.EmbedAllRequest) (*entities.EmbedAllRequest, error) {
	domainReq := &entities.EmbedAllRequest{
		Inputs: entities.Input{Data: req.Inputs},
	}

	if req.PromptName != nil {
		domainReq.PromptName = req.PromptName
	}
	if req.Truncate != nil {
		domainReq.Truncate = req.Truncate
	}
	if req.TruncationDirection != nil {
		domainReq.TruncationDirection = convertTruncationDirection(*req.TruncationDirection)
	}

	return domainReq, nil
}

func (s *Server) convertEmbedSparseRequest(req *pb.EmbedSparseRequest) (*entities.EmbedSparseRequest, error) {
	domainReq := &entities.EmbedSparseRequest{
		Inputs: entities.Input{Data: req.Inputs},
	}

	if req.PromptName != nil {
		domainReq.PromptName = req.PromptName
	}
	if req.Truncate != nil {
		domainReq.Truncate = req.Truncate
	}
	if req.TruncationDirection != nil {
		domainReq.TruncationDirection = convertTruncationDirection(*req.TruncationDirection)
	}

	return domainReq, nil
}

func (s *Server) convertSimilarityRequest(req *pb.SimilarityRequest) (*entities.SimilarityRequest, error) {
	domainReq := &entities.SimilarityRequest{
		Inputs: entities.SimilarityInput{
			SourceSentence: req.SourceSentence,
			Sentences:      req.Sentences,
		},
	}

	if req.Parameters != nil {
		domainReq.Parameters = &entities.SimilarityParameters{}
		if req.Parameters.PromptName != nil {
			domainReq.Parameters.PromptName = req.Parameters.PromptName
		}
		if req.Parameters.Truncate != nil {
			domainReq.Parameters.Truncate = req.Parameters.Truncate
		}
		if req.Parameters.TruncationDirection != nil {
			domainReq.Parameters.TruncationDirection = convertTruncationDirection(*req.Parameters.TruncationDirection)
		}
	}

	return domainReq, nil
}

// Convert domain responses to protobuf responses

func (s *Server) convertEmbedResponse(resp *entities.EmbedResponse) *pb.EmbedResponse {
	embeddings := make([]*pb.Embedding, len(resp.Embeddings))
	for i, embedding := range resp.Embeddings {
		embeddings[i] = &pb.Embedding{Values: embedding}
	}
	return &pb.EmbedResponse{Embeddings: embeddings}
}

func (s *Server) convertEmbedAllResponse(resp *entities.EmbedAllResponse) *pb.EmbedAllResponse {
	tokenEmbeddings := make([]*pb.TokenEmbeddings, len(resp.Embeddings))
	for i, textEmbeddings := range resp.Embeddings {
		embeddings := make([]*pb.Embedding, len(textEmbeddings))
		for j, embedding := range textEmbeddings {
			embeddings[j] = &pb.Embedding{Values: embedding}
		}
		tokenEmbeddings[i] = &pb.TokenEmbeddings{Embeddings: embeddings}
	}
	return &pb.EmbedAllResponse{TokenEmbeddings: tokenEmbeddings}
}

func (s *Server) convertEmbedSparseResponse(resp *entities.EmbedSparseResponse) *pb.EmbedSparseResponse {
	sparseEmbeddings := make([]*pb.SparseEmbedding, len(resp.Embeddings))
	for i, embedding := range resp.Embeddings {
		values := make([]*pb.SparseValue, len(embedding))
		for j, val := range embedding {
			values[j] = &pb.SparseValue{
				Index: uint32(val.Index),
				Value: val.Value,
			}
		}
		sparseEmbeddings[i] = &pb.SparseEmbedding{Values: values}
	}
	return &pb.EmbedSparseResponse{SparseEmbeddings: sparseEmbeddings}
}

// Helper conversion functions

func convertTruncationDirection(dir pb.TruncationDirection) entities.TruncationDirection {
	switch dir {
	case pb.TruncationDirection_TRUNCATION_DIRECTION_LEFT:
		return entities.TruncationLeft
	case pb.TruncationDirection_TRUNCATION_DIRECTION_RIGHT:
		return entities.TruncationRight
	default:
		return entities.TruncationRight
	}
}

// func convertEncodingFormat(format pb.EncodingFormat) entities.EncodingFormat {
// 	switch format {
// 	case pb.EncodingFormat_ENCODING_FORMAT_FLOAT:
// 		return entities.EncodingFloat
// 	case pb.EncodingFormat_ENCODING_FORMAT_BASE64:
// 		return entities.EncodingBase64
// 	default:
// 		return entities.EncodingFloat
// 	}
// }

// Error conversion

func (s *Server) convertError(err error) error {
	if teiErr, ok := err.(*errors.TEIError); ok {
		return s.convertTEIError(teiErr)
	}

	if validationErr, ok := err.(*errors.ValidationError); ok {
		return status.Errorf(codes.InvalidArgument, "validation error: %s", validationErr.Message)
	}

	if multiValidationErr, ok := err.(*errors.MultiValidationError); ok {
		return status.Errorf(codes.InvalidArgument, "validation errors: %s", multiValidationErr.Error())
	}

	// Generic error
	return status.Errorf(codes.Internal, "internal error: %v", err)
}

func (s *Server) convertTEIError(teiErr *errors.TEIError) error {
	var code codes.Code

	switch teiErr.Type {
	case errors.ErrorTypeValidation:
		code = codes.InvalidArgument
	case errors.ErrorTypeTokenizer:
		code = codes.InvalidArgument
	case errors.ErrorTypeBackend:
		code = codes.Internal
	case errors.ErrorTypeOverloaded:
		code = codes.ResourceExhausted
	case errors.ErrorTypeUnhealthy:
		code = codes.Unavailable
	case errors.ErrorTypeNetwork:
		code = codes.Unavailable
	case errors.ErrorTypeTimeout:
		code = codes.DeadlineExceeded
	default:
		code = codes.Internal
	}

	return status.Errorf(code, "[%s] %s", teiErr.Type, teiErr.Message)
}
