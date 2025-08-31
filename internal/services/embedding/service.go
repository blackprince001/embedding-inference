package embedding

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blackprince001/embedding-inference/internal/domain/entities"
	"github.com/blackprince001/embedding-inference/internal/domain/errors"
	"github.com/blackprince001/embedding-inference/internal/domain/interfaces"

	"go.uber.org/zap"
)

type Service struct {
	httpClient interfaces.HTTPClient
	logger     *zap.Logger
	validator  *entities.Validator
}

func NewService(httpClient interfaces.HTTPClient, logger *zap.Logger) *Service {
	return &Service{
		httpClient: httpClient,
		logger:     logger.Named("embedding"),
		validator:  entities.NewValidator(entities.DefaultValidationConfig()),
	}
}

func (s *Service) Embed(ctx context.Context, req *entities.EmbedRequest) (*entities.EmbedResponse, error) {
	s.logger.Debug("Processing embed request",
		zap.Int("input_count", len(req.Inputs.Data)),
		zap.Bool("normalize", req.Normalize != nil && *req.Normalize),
	)

	req.SetDefaults()

	if err := s.validator.ValidateEmbedRequest(req); err != nil {
		s.logger.Error("Embed request validation failed", zap.Error(err))
		return nil, err
	}

	responseData, err := s.httpClient.Post(ctx, entities.EndpointEmbed, req)
	if err != nil {
		s.logger.Error("Embed request failed", zap.Error(err))
		return nil, fmt.Errorf("embed request failed: %w", err)
	}

	var response entities.EmbedResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		s.logger.Error("Failed to parse embed response", zap.Error(err))
		return nil, errors.NewTEIError("failed to parse response", errors.ErrorTypeBackend)
	}

	s.logger.Debug("Embed request completed",
		zap.Int("embeddings_count", len(response.Embeddings)),
	)

	return &response, nil
}

func (s *Service) EmbedAll(ctx context.Context, req *entities.EmbedAllRequest) (*entities.EmbedAllResponse, error) {
	s.logger.Debug("Processing embed_all request",
		zap.Int("input_count", len(req.Inputs.Data)),
	)

	req.SetDefaults()

	if err := req.Validate(); err != nil {
		s.logger.Error("EmbedAll request validation failed", zap.Error(err))
		return nil, err
	}

	responseData, err := s.httpClient.Post(ctx, entities.EndpointEmbedAll, req)
	if err != nil {
		s.logger.Error("EmbedAll request failed", zap.Error(err))
		return nil, fmt.Errorf("embed_all request failed: %w", err)
	}

	var response entities.EmbedAllResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		s.logger.Error("Failed to parse embed_all response", zap.Error(err))
		return nil, errors.NewTEIError("failed to parse response", errors.ErrorTypeBackend)
	}

	s.logger.Debug("EmbedAll request completed",
		zap.Int("embeddings_count", len(response.Embeddings)),
	)

	return &response, nil
}

func (s *Service) EmbedSparse(ctx context.Context, req *entities.EmbedSparseRequest) (*entities.EmbedSparseResponse, error) {
	s.logger.Debug("Processing embed_sparse request",
		zap.Int("input_count", len(req.Inputs.Data)),
	)

	req.SetDefaults()

	if err := req.Validate(); err != nil {
		s.logger.Error("EmbedSparse request validation failed", zap.Error(err))
		return nil, err
	}

	responseData, err := s.httpClient.Post(ctx, entities.EndpointEmbedSparse, req)
	if err != nil {
		s.logger.Error("EmbedSparse request failed", zap.Error(err))
		return nil, fmt.Errorf("embed_sparse request failed: %w", err)
	}

	var response entities.EmbedSparseResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		s.logger.Error("Failed to parse embed_sparse response", zap.Error(err))
		return nil, errors.NewTEIError("failed to parse response", errors.ErrorTypeBackend)
	}

	s.logger.Debug("EmbedSparse request completed",
		zap.Int("embeddings_count", len(response.Embeddings)),
	)

	return &response, nil
}
