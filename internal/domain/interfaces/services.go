package interfaces

import (
	"context"
	"time"

	"github.com/blackprince001/embedding-inference/internal/domain/entities"
)

type EmbeddingService interface {
	Embed(ctx context.Context, req *entities.EmbedRequest) (*entities.EmbedResponse, error)
	EmbedAll(ctx context.Context, req *entities.EmbedAllRequest) (*entities.EmbedAllResponse, error)
	EmbedSparse(ctx context.Context, req *entities.EmbedSparseRequest) (*entities.EmbedSparseResponse, error)
}

type SimilarityService interface {
	CalculateSimilarity(ctx context.Context, req *entities.SimilarityRequest) (*entities.SimilarityResponse, error)
}

type ClientService interface {
	EmbeddingService
	SimilarityService
}

type HTTPClient interface {
	Get(ctx context.Context, endpoint string) ([]byte, error)
	Post(ctx context.Context, endpoint string, body any) ([]byte, error)
	PostRaw(ctx context.Context, endpoint string, body []byte, contentType string) ([]byte, error)
	SetTimeout(timeout time.Duration)
	Close() error
}
