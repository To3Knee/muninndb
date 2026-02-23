package embed

import (
	"context"
	"testing"
)

// MockProvider is a mock implementation of the Provider interface for testing.
type MockProvider struct {
	maxBatchSize int
	callCount    int
	lastTexts    []string
}

func (m *MockProvider) Name() string {
	return "mock"
}

func (m *MockProvider) Init(ctx context.Context, cfg ProviderHTTPConfig) (int, error) {
	return 2, nil
}

func (m *MockProvider) EmbedBatch(ctx context.Context, texts []string) ([]float32, error) {
	m.callCount++
	m.lastTexts = texts

	// Return dummy embeddings: one per text, 2 dimensions each
	result := make([]float32, len(texts)*2)
	for i := 0; i < len(texts)*2; i++ {
		result[i] = float32(i) / 10.0
	}
	return result, nil
}

func (m *MockProvider) MaxBatchSize() int {
	return m.maxBatchSize
}

func (m *MockProvider) Close() error {
	return nil
}

func TestBatchEmbedder_SingleBatch(t *testing.T) {
	mock := &MockProvider{maxBatchSize: 32}
	embedder := NewBatchEmbedder(mock, nil)

	texts := []string{"hello", "world"}
	result, err := embedder.Embed(context.Background(), texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if mock.callCount != 1 {
		t.Errorf("expected 1 provider call, got %d", mock.callCount)
	}

	expectedLen := len(texts) * 2 // 2 texts * 2 dimensions
	if len(result) != expectedLen {
		t.Errorf("expected %d embeddings, got %d", expectedLen, len(result))
	}
}

func TestBatchEmbedder_MultipleBatches(t *testing.T) {
	mock := &MockProvider{maxBatchSize: 2}
	embedder := NewBatchEmbedder(mock, nil)

	texts := []string{"a", "b", "c", "d", "e"}
	result, err := embedder.Embed(context.Background(), texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	// With batch size 2 and 5 texts, expect 3 calls (2, 2, 1)
	expectedCalls := 3
	if mock.callCount != expectedCalls {
		t.Errorf("expected %d provider calls, got %d", expectedCalls, mock.callCount)
	}

	expectedLen := len(texts) * 2 // 5 texts * 2 dimensions
	if len(result) != expectedLen {
		t.Errorf("expected %d embeddings, got %d", expectedLen, len(result))
	}
}

func TestBatchEmbedder_ExactBatchSize(t *testing.T) {
	mock := &MockProvider{maxBatchSize: 3}
	embedder := NewBatchEmbedder(mock, nil)

	texts := []string{"a", "b", "c", "d", "e", "f"}
	result, err := embedder.Embed(context.Background(), texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	// With batch size 3 and 6 texts, expect 2 calls
	expectedCalls := 2
	if mock.callCount != expectedCalls {
		t.Errorf("expected %d provider calls, got %d", expectedCalls, mock.callCount)
	}

	expectedLen := len(texts) * 2
	if len(result) != expectedLen {
		t.Errorf("expected %d embeddings, got %d", expectedLen, len(result))
	}
}

func TestBatchEmbedder_SingleText(t *testing.T) {
	mock := &MockProvider{maxBatchSize: 10}
	embedder := NewBatchEmbedder(mock, nil)

	texts := []string{"single"}
	result, err := embedder.Embed(context.Background(), texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if mock.callCount != 1 {
		t.Errorf("expected 1 provider call, got %d", mock.callCount)
	}

	expectedLen := 2 // 1 text * 2 dimensions
	if len(result) != expectedLen {
		t.Errorf("expected %d embeddings, got %d", expectedLen, len(result))
	}
}

func TestBatchEmbedder_EmptyInput(t *testing.T) {
	mock := &MockProvider{maxBatchSize: 32}
	embedder := NewBatchEmbedder(mock, nil)

	texts := []string{}
	result, err := embedder.Embed(context.Background(), texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if mock.callCount != 0 {
		t.Errorf("expected 0 provider calls, got %d", mock.callCount)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 embeddings, got %d", len(result))
	}
}

func TestBatchEmbedder_LargeBatch(t *testing.T) {
	mock := &MockProvider{maxBatchSize: 10}
	embedder := NewBatchEmbedder(mock, nil)

	// Create 25 texts
	texts := make([]string, 25)
	for i := 0; i < 25; i++ {
		texts[i] = "text"
	}

	result, err := embedder.Embed(context.Background(), texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	// With batch size 10 and 25 texts, expect 3 calls (10, 10, 5)
	expectedCalls := 3
	if mock.callCount != expectedCalls {
		t.Errorf("expected %d provider calls, got %d", expectedCalls, mock.callCount)
	}

	expectedLen := len(texts) * 2 // 25 texts * 2 dimensions
	if len(result) != expectedLen {
		t.Errorf("expected %d embeddings, got %d", expectedLen, len(result))
	}
}

func TestBatchEmbedder_ContextCancellation(t *testing.T) {
	mock := &MockProvider{maxBatchSize: 2}
	embedder := NewBatchEmbedder(mock, nil)

	texts := []string{"a", "b", "c", "d"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This should fail or succeed quickly due to cancelled context
	// (no rate limiter, so it goes straight to provider)
	_, err := embedder.Embed(ctx, texts)
	if err != nil && err != context.Canceled {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBatchEmbedder_MockEmbedding(t *testing.T) {
	mock := &MockProvider{maxBatchSize: 2}
	embedder := NewBatchEmbedder(mock, nil)

	texts := []string{"hello", "world"}
	result, err := embedder.Embed(context.Background(), texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	// Verify mock was called with correct texts
	if len(mock.lastTexts) != len(texts) {
		t.Errorf("expected mock to be called with %d texts, got %d", len(texts), len(mock.lastTexts))
	}

	for i, text := range texts {
		if mock.lastTexts[i] != text {
			t.Errorf("expected text %q, got %q", text, mock.lastTexts[i])
		}
	}

	// Verify result structure
	if len(result) != len(texts)*2 {
		t.Errorf("expected %d values, got %d", len(texts)*2, len(result))
	}
}
