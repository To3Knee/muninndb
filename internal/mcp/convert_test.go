package mcp

import (
	"testing"

	"github.com/scrypster/muninndb/internal/transport/mbp"
)

func TestConvertActivationToMemory(t *testing.T) {
	item := &mbp.ActivationItem{
		ID:         "abc123",
		Concept:    "test concept",
		Content:    "short content",
		Score:      0.9,
		Confidence: 0.85,
		Why:        "found in context",
	}
	m := activationToMemory(item)
	if m.Concept != "test concept" {
		t.Errorf("concept = %q, want %q", m.Concept, "test concept")
	}
	if m.Content != "short content" {
		t.Errorf("content = %q, want %q", m.Content, "short content")
	}
	if m.ID != "abc123" {
		t.Errorf("id = %q, want %q", m.ID, "abc123")
	}
}

func TestConvertTruncatesLongContent(t *testing.T) {
	long := make([]byte, 600)
	for i := range long {
		long[i] = 'x'
	}

	item := &mbp.ActivationItem{
		ID:      "test-id",
		Content: string(long),
	}
	m := activationToMemory(item)
	if len(m.Content) > 503 { // 500 + "..."
		t.Errorf("content not truncated: len=%d", len(m.Content))
	}
	if m.Content[len(m.Content)-3:] != "..." {
		t.Error("truncated content must end with '...'")
	}
}

func TestConvertUsesContentWhenNoSummary(t *testing.T) {
	item := &mbp.ActivationItem{
		ID:      "test-id",
		Content: "the content",
	}
	m := activationToMemory(item)
	if m.Content != "the content" {
		t.Errorf("content = %q, want %q", m.Content, "the content")
	}
}

func TestConvertReadResponseToMemory(t *testing.T) {
	resp := &mbp.ReadResponse{
		ID:         "read-123",
		Concept:    "stored concept",
		Content:    "stored content",
		Confidence: 0.95,
		State:      1,
		Tags:       []string{"tag1", "tag2"},
	}
	m := readResponseToMemory(resp)
	if m.ID != "read-123" {
		t.Errorf("id = %q, want %q", m.ID, "read-123")
	}
	if m.Concept != "stored concept" {
		t.Errorf("concept = %q, want %q", m.Concept, "stored concept")
	}
	if len(m.Tags) != 2 {
		t.Errorf("tags len = %d, want 2", len(m.Tags))
	}
}
