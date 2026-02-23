package embed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scrypster/muninndb/internal/plugin"
)

func TestNewEmbedService_Ollama(t *testing.T) {
	es, err := NewEmbedService("ollama://localhost:11434/nomic-embed-text")
	if err != nil {
		t.Fatalf("NewEmbedService failed: %v", err)
	}

	if es.name != "embed-ollama" {
		t.Errorf("expected name embed-ollama, got %s", es.name)
	}

	if es.provCfg.Model != "nomic-embed-text" {
		t.Errorf("expected model nomic-embed-text, got %s", es.provCfg.Model)
	}
}

func TestNewEmbedService_OpenAI(t *testing.T) {
	es, err := NewEmbedService("openai://text-embedding-3-small")
	if err != nil {
		t.Fatalf("NewEmbedService failed: %v", err)
	}

	if es.name != "embed-openai" {
		t.Errorf("expected name embed-openai, got %s", es.name)
	}

	if es.provCfg.Model != "text-embedding-3-small" {
		t.Errorf("expected model text-embedding-3-small, got %s", es.provCfg.Model)
	}
}

func TestNewEmbedService_Voyage(t *testing.T) {
	es, err := NewEmbedService("voyage://voyage-3")
	if err != nil {
		t.Fatalf("NewEmbedService failed: %v", err)
	}

	if es.name != "embed-voyage" {
		t.Errorf("expected name embed-voyage, got %s", es.name)
	}

	if es.provCfg.Model != "voyage-3" {
		t.Errorf("expected model voyage-3, got %s", es.provCfg.Model)
	}
}

func TestNewEmbedService_InvalidScheme(t *testing.T) {
	_, err := NewEmbedService("unknown://model")
	if err == nil {
		t.Fatal("expected error for invalid scheme")
	}
}

func TestEmbedServiceName(t *testing.T) {
	es, _ := NewEmbedService("ollama://localhost:11434/model")
	if es.Name() != "embed-ollama" {
		t.Errorf("expected embed-ollama, got %s", es.Name())
	}
}

func TestEmbedServiceTier(t *testing.T) {
	es, _ := NewEmbedService("ollama://localhost:11434/model")
	if es.Tier() != plugin.TierEmbed {
		t.Errorf("expected TierEmbed, got %v", es.Tier())
	}
}

func TestEmbedServiceDimension(t *testing.T) {
	es := &EmbedService{dim: 1536}
	if es.Dimension() != 1536 {
		t.Errorf("expected dimension 1536, got %d", es.Dimension())
	}
}

func TestEmbedService_OllamaInit_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "POST" && r.URL.Path == "/api/embeddings" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"embedding": [0.1, 0.2, 0.3, 0.4]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	es, _ := NewEmbedService("ollama://" + server.Listener.Addr().String() + "/test-model")

	cfg := plugin.PluginConfig{
		ProviderURL: "ollama://" + server.Listener.Addr().String() + "/test-model",
		APIKey:      "",
		Options:     map[string]string{},
	}

	err := es.Init(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if es.Dimension() != 4 {
		t.Errorf("expected dimension 4, got %d", es.Dimension())
	}
}

func TestEmbedService_OllamaInit_Unreachable(t *testing.T) {
	es, _ := NewEmbedService("ollama://localhost:54321/test-model")

	cfg := plugin.PluginConfig{
		ProviderURL: "ollama://localhost:54321/test-model",
	}

	err := es.Init(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for unreachable Ollama")
	}
}

func TestEmbedService_EmbedAfterInit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "POST" && r.URL.Path == "/api/embeddings" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"embedding": [0.1, 0.2]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	es, _ := NewEmbedService("ollama://" + server.Listener.Addr().String() + "/test-model")

	cfg := plugin.PluginConfig{
		ProviderURL: "ollama://" + server.Listener.Addr().String() + "/test-model",
	}

	es.Init(context.Background(), cfg)

	result, err := es.Embed(context.Background(), []string{"hello world"})
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	expectedLen := 2 // 1 text * 2 dimension
	if len(result) != expectedLen {
		t.Errorf("expected %d embeddings, got %d", expectedLen, len(result))
	}
}

func TestEmbedService_Close(t *testing.T) {
	es, _ := NewEmbedService("ollama://localhost:11434/model")

	err := es.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Second close should be no-op
	err = es.Close()
	if err != nil {
		t.Errorf("second Close failed: %v", err)
	}
}
