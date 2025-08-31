package client

import (
	"context"
	"fmt"

	"teiwrappergolang/internal/config"
	"teiwrappergolang/internal/domain/entities"
	"teiwrappergolang/internal/domain/interfaces"
	"teiwrappergolang/internal/infrastructure/logging"
	"teiwrappergolang/internal/services/embedding"
	"teiwrappergolang/internal/services/similarity"
)

type Client struct {
	embeddingService  interfaces.EmbeddingService
	similarityService interfaces.SimilarityService
	httpClient        interfaces.HTTPClient

	config *config.Config
	logger *logging.Logger
}

func NewClient(cfg *config.Config, httpClient interfaces.HTTPClient, logger *logging.Logger) *Client {
	clientLogger := logger.Named("tei-client")

	return &Client{
		embeddingService:  embedding.NewService(httpClient, clientLogger),
		similarityService: similarity.NewService(httpClient, clientLogger),
		httpClient:        httpClient,
		config:            cfg,
		logger:            logger,
	}
}

func (c *Client) Embed(ctx context.Context, req *entities.EmbedRequest) (*entities.EmbedResponse, error) {
	return c.embeddingService.Embed(ctx, req)
}

func (c *Client) EmbedAll(ctx context.Context, req *entities.EmbedAllRequest) (*entities.EmbedAllResponse, error) {
	return c.embeddingService.EmbedAll(ctx, req)
}

func (c *Client) EmbedSparse(ctx context.Context, req *entities.EmbedSparseRequest) (*entities.EmbedSparseResponse, error) {
	return c.embeddingService.EmbedSparse(ctx, req)
}

func (c *Client) CalculateSimilarity(ctx context.Context, req *entities.SimilarityRequest) (*entities.SimilarityResponse, error) {
	return c.similarityService.CalculateSimilarity(ctx, req)
}

func (c *Client) EmbedTexts(ctx context.Context, texts []string, normalize bool) (*entities.EmbedResponse, error) {
	req := &entities.EmbedRequest{
		Inputs:    entities.Input{Data: texts},
		Normalize: &normalize,
	}
	return c.Embed(ctx, req)
}

func (c *Client) EmbedText(ctx context.Context, text string, normalize bool) ([]float32, error) {
	resp, err := c.EmbedTexts(ctx, []string{text}, normalize)
	if err != nil {
		return nil, err
	}
	if len(resp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return resp.Embeddings[0], nil
}

func (c *Client) CalculateTextSimilarity(ctx context.Context, source string, targets []string) ([]float32, error) {
	req := &entities.SimilarityRequest{
		Inputs: entities.SimilarityInput{
			SourceSentence: source,
			Sentences:      targets,
		},
	}
	resp, err := c.CalculateSimilarity(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Similarities, nil
}
