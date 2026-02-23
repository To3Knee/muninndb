package enrich

import (
	"testing"

	"github.com/scrypster/muninndb/internal/plugin"
)

// TestEnrichServiceNew_Ollama tests creating an EnrichService for Ollama.
func TestEnrichServiceNew_Ollama(t *testing.T) {
	es, err := NewEnrichService("ollama://localhost:11434/llama3.2")
	if err != nil {
		t.Fatalf("NewEnrichService failed: %v", err)
	}

	if es.Name() != "enrich-ollama" {
		t.Fatalf("Expected name 'enrich-ollama', got: %q", es.Name())
	}

	if es.Tier() != plugin.TierEnrich {
		t.Fatalf("Expected tier TierEnrich (3), got: %d", es.Tier())
	}

	if es.provCfg.Model != "llama3.2" {
		t.Fatalf("Expected model 'llama3.2', got: %q", es.provCfg.Model)
	}

	_ = es.Close()
}

// TestEnrichServiceNew_OpenAI tests creating an EnrichService for OpenAI.
func TestEnrichServiceNew_OpenAI(t *testing.T) {
	es, err := NewEnrichService("openai://gpt-4o-mini")
	if err != nil {
		t.Fatalf("NewEnrichService failed: %v", err)
	}

	if es.Name() != "enrich-openai" {
		t.Fatalf("Expected name 'enrich-openai', got: %q", es.Name())
	}

	if es.Tier() != plugin.TierEnrich {
		t.Fatalf("Expected tier TierEnrich (3), got: %d", es.Tier())
	}

	if es.provCfg.Model != "gpt-4o-mini" {
		t.Fatalf("Expected model 'gpt-4o-mini', got: %q", es.provCfg.Model)
	}

	_ = es.Close()
}

// TestEnrichServiceNew_Anthropic tests creating an EnrichService for Anthropic.
func TestEnrichServiceNew_Anthropic(t *testing.T) {
	es, err := NewEnrichService("anthropic://claude-haiku")
	if err != nil {
		t.Fatalf("NewEnrichService failed: %v", err)
	}

	if es.Name() != "enrich-anthropic" {
		t.Fatalf("Expected name 'enrich-anthropic', got: %q", es.Name())
	}

	if es.Tier() != plugin.TierEnrich {
		t.Fatalf("Expected tier TierEnrich (3), got: %d", es.Tier())
	}

	if es.provCfg.Model != "claude-haiku" {
		t.Fatalf("Expected model 'claude-haiku', got: %q", es.provCfg.Model)
	}

	_ = es.Close()
}

// TestEnrichServiceNew_InvalidScheme tests error handling for invalid schemes.
func TestEnrichServiceNew_InvalidScheme(t *testing.T) {
	_, err := NewEnrichService("invalid://localhost:11434/model")
	if err == nil {
		t.Fatalf("Expected error for invalid scheme, got nil")
	}
}
