package embed

import (
	"context"
	"fmt"
)

// BatchEmbedder splits large text arrays into chunks before sending to provider.
type BatchEmbedder struct {
	provider  Provider
	limiter   *TokenBucketLimiter
	batchSize int // max texts per request
}

// NewBatchEmbedder creates a new BatchEmbedder.
func NewBatchEmbedder(provider Provider, limiter *TokenBucketLimiter) *BatchEmbedder {
	return &BatchEmbedder{
		provider:  provider,
		limiter:   limiter,
		batchSize: provider.MaxBatchSize(),
	}
}

// Embed sends texts in batches, returns concatenated embeddings.
func (b *BatchEmbedder) Embed(ctx context.Context, texts []string) ([]float32, error) {
	result := make([]float32, 0)

	for i := 0; i < len(texts); i += b.batchSize {
		end := i + b.batchSize
		if end > len(texts) {
			end = len(texts)
		}

		// Wait for rate limit token if limiter exists
		if b.limiter != nil {
			if err := b.limiter.Wait(ctx); err != nil {
				return nil, err
			}
		}

		chunk, err := b.provider.EmbedBatch(ctx, texts[i:end])
		if err != nil {
			return nil, err
		}
		if len(chunk) == 0 {
			return nil, fmt.Errorf("embed: provider returned no embeddings for batch")
		}
		result = append(result, chunk...)
	}

	return result, nil
}
