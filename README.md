# Text Embeddings Client

A Go wrapper service for Hugging Face Text Embeddings Inference (TEI) API, designed with clean architecture principles and supporting both HTTP client library and gRPC microservice interfaces.

## Features

- **Embedding Operations**: Support for standard, sparse, and OpenAI-compatible embeddings
- **Similarity Calculations**: Sentence similarity computations with ranking

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository and set your model in the config file.

2. Start all services:

   ```bash
   docker-compose up -d
   ```

### Local Development

1. Install dependencies with from golang mod file along with protobuf requirements.

2. Start TEI service:

   ```bash
   docker run -p 8080:8080 \
     ghcr.io/huggingface/text-embeddings-inference:1.5 \
     --model-id thenlper/gte-base
   ```

3. Run HTTP client by building the project.

## gRPC API

### Service Definition

```protobuf
service TextEmbeddingsService {
  // Embedding operations
  rpc Embed(EmbedRequest) returns (EmbedResponse);
  rpc EmbedAll(EmbedAllRequest) returns (EmbedAllResponse);
  rpc EmbedSparse(EmbedSparseRequest) returns (EmbedSparseResponse);
  
  // Similarity operations
  rpc CalculateSimilarity(SimilarityRequest) returns (SimilarityResponse);
}
```

### Testing gRPC API

```bash
# List available services
grpcurl -plaintext localhost:9090 list

# Get service info
grpcurl -plaintext localhost:9090 text_embeddings.v1.TextEmbeddingsService.GetInfo

# Test embedding
grpcurl -plaintext -d '{"inputs":["Hello world"]}' \
  localhost:9090 text_embeddings.v1.TextEmbeddingsService.Embed
```

## Configuration

- `configs/config.yaml`: Default configuration
- `configs/docker.yaml`: Docker-specific configuration

## Client Library Usage

### HTTP Client

```go
import "text-embeddings-client/pkg/client"

// Initialize
registry := services.NewServiceRegistry(config, logger)
registry.Initialize()
client := registry.GetClient()
client.Connect(ctx)

// Embed text
embeddings, err := client.EmbedTexts(ctx, []string{"Hello world"}, true)

// Calculate similarity
similarities, err := client.CalculateTextSimilarity(ctx, "source", targets)
```

### gRPC Client

```go
import pb "text-embeddings-client/api/gen/text_embeddings/v1"

// Connect
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
client := pb.NewTextEmbeddingsServiceClient(conn)

// Embed
resp, err := client.Embed(ctx, &pb.EmbedRequest{
    Inputs: []string{"Hello world"},
    Normalize: &[]bool{true}[0],
})
```
