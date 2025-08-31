package similarity

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
		logger:     logger.Named("similarity"),
		validator:  entities.NewValidator(entities.DefaultValidationConfig()),
	}
}

func (s *Service) CalculateSimilarity(ctx context.Context, req *entities.SimilarityRequest) (*entities.SimilarityResponse, error) {
	s.logger.Debug("Processing similarity request",
		zap.String("source_sentence", truncateString(req.Inputs.SourceSentence, 50)),
		zap.Int("sentences_count", len(req.Inputs.Sentences)),
	)

	req.SetDefaults()

	if err := s.validator.ValidateSimilarityRequest(req); err != nil {
		s.logger.Error("Similarity request validation failed", zap.Error(err))
		return nil, err
	}

	responseData, err := s.httpClient.Post(ctx, entities.EndpointSimilarity, req)
	if err != nil {
		s.logger.Error("Similarity request failed", zap.Error(err))
		return nil, fmt.Errorf("similarity request failed: %w", err)
	}

	var response entities.SimilarityResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		s.logger.Error("Failed to parse similarity response", zap.Error(err))
		return nil, errors.NewTEIError("failed to parse response", errors.ErrorTypeBackend)
	}

	if len(response.Similarities) != len(req.Inputs.Sentences) {
		s.logger.Error("Response similarity count mismatch",
			zap.Int("expected", len(req.Inputs.Sentences)),
			zap.Int("received", len(response.Similarities)),
		)
		return nil, errors.NewTEIError("response similarity count mismatch", errors.ErrorTypeBackend)
	}

	s.logger.Debug("Similarity request completed",
		zap.Int("similarities_count", len(response.Similarities)),
		zap.Float32("avg_similarity", calculateAverage(response.Similarities)),
	)

	return &response, nil
}

func (s *Service) CalculatePairwiseSimilarity(ctx context.Context, sentences1, sentences2 []string) ([][]float32, error) {
	if len(sentences1) == 0 || len(sentences2) == 0 {
		return nil, errors.NewValidationError("sentences", "both sentence arrays must be non-empty", nil)
	}

	results := make([][]float32, len(sentences1))

	for i, sentence1 := range sentences1 {
		req := &entities.SimilarityRequest{
			Inputs: entities.SimilarityInput{
				SourceSentence: sentence1,
				Sentences:      sentences2,
			},
		}

		resp, err := s.CalculateSimilarity(ctx, req)
		if err != nil {
			s.logger.Error("Pairwise similarity calculation failed",
				zap.Int("sentence1_index", i),
				zap.Error(err),
			)
			return nil, fmt.Errorf("pairwise similarity failed at index %d: %w", i, err)
		}

		results[i] = resp.Similarities
	}

	s.logger.Debug("Pairwise similarity completed",
		zap.Int("sentences1_count", len(sentences1)),
		zap.Int("sentences2_count", len(sentences2)),
	)

	return results, nil
}

func (s *Service) FindMostSimilar(ctx context.Context, sourceSentence string, candidates []string, topK int) (*MostSimilarResult, error) {
	if topK <= 0 {
		return nil, errors.NewValidationError("topK", "must be positive", topK)
	}

	if topK > len(candidates) {
		topK = len(candidates)
	}

	req := &entities.SimilarityRequest{
		Inputs: entities.SimilarityInput{
			SourceSentence: sourceSentence,
			Sentences:      candidates,
		},
	}

	resp, err := s.CalculateSimilarity(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("similarity calculation failed: %w", err)
	}

	type indexedSimilarity struct {
		Index      int
		Sentence   string
		Similarity float32
	}

	indexed := make([]indexedSimilarity, len(resp.Similarities))
	for i, sim := range resp.Similarities {
		indexed[i] = indexedSimilarity{
			Index:      i,
			Sentence:   candidates[i],
			Similarity: sim,
		}
	}

	for i := 0; i < len(indexed); i++ {
		for j := i + 1; j < len(indexed); j++ {
			if indexed[j].Similarity > indexed[i].Similarity {
				indexed[i], indexed[j] = indexed[j], indexed[i]
			}
		}
	}

	results := make([]SimilarSentence, topK)
	for i := 0; i < topK; i++ {
		results[i] = SimilarSentence{
			Index:      indexed[i].Index,
			Sentence:   indexed[i].Sentence,
			Similarity: indexed[i].Similarity,
		}
	}

	return &MostSimilarResult{
		SourceSentence: sourceSentence,
		TopMatches:     results,
	}, nil
}

type MostSimilarResult struct {
	SourceSentence string            `json:"source_sentence"`
	TopMatches     []SimilarSentence `json:"top_matches"`
}

type SimilarSentence struct {
	Index      int     `json:"index"`
	Sentence   string  `json:"sentence"`
	Similarity float32 `json:"similarity"`
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func calculateAverage(values []float32) float32 {
	if len(values) == 0 {
		return 0.0
	}

	var sum float32
	for _, v := range values {
		sum += v
	}

	return sum / float32(len(values))
}
