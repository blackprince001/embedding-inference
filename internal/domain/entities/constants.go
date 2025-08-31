package entities

import "time"

const (
	EndpointEmbed       = "/embed"
	EndpointEmbedAll    = "/embed_all"
	EndpointEmbedSparse = "/embed_sparse"
	EndpointEmbedOpenAI = "/v1/embeddings"
	EndpointSimilarity  = "/similarity"
	EndpointTokenize    = "/tokenize"
	EndpointDecode      = "/decode"
)

const (
	HeaderContentType   = "Content-Type"
	HeaderAccept        = "Accept"
	HeaderUserAgent     = "User-Agent"
	HeaderAuthorization = "Authorization"
	HeaderRequestID     = "X-Request-ID"
)

const (
	ContentTypeJSON       = "application/json"
	ContentTypeTextPlain  = "text/plain"
	ContentTypePrometheus = "text/plain; version=0.0.4"
)

const (
	DefaultTimeout           = 30 * time.Second
	DefaultMaxRetries        = 3
	DefaultRetryDelay        = 1 * time.Second
	DefaultMaxConnections    = 10
	DefaultMaxInputLength    = 8192
	DefaultMaxBatchSize      = 32
	DefaultMaxSentencesCount = 100
	DefaultNormalize         = true
	DefaultTruncate          = false
	DefaultAddSpecialTokens  = true
	DefaultSkipSpecialTokens = true
)

const (
	StatusOK                    = 200
	StatusBadRequest            = 400
	StatusRequestEntityTooLarge = 413
	StatusUnprocessableEntity   = 422
	StatusFailedDependency      = 424
	StatusTooManyRequests       = 429
	StatusInternalServerError   = 500
	StatusServiceUnavailable    = 503
)
