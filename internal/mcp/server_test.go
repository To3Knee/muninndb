package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/scrypster/muninndb/internal/transport/mbp"
)

// fakeEngine implements EngineInterface for tests.
type fakeEngine struct{}

func (f *fakeEngine) Write(ctx context.Context, req *mbp.WriteRequest) (*mbp.WriteResponse, error) {
	return &mbp.WriteResponse{}, nil
}
func (f *fakeEngine) Activate(ctx context.Context, req *mbp.ActivateRequest) (*mbp.ActivateResponse, error) {
	return &mbp.ActivateResponse{}, nil
}
func (f *fakeEngine) Read(ctx context.Context, req *mbp.ReadRequest) (*mbp.ReadResponse, error) {
	return &mbp.ReadResponse{}, nil
}
func (f *fakeEngine) Forget(ctx context.Context, req *mbp.ForgetRequest) (*mbp.ForgetResponse, error) {
	return &mbp.ForgetResponse{}, nil
}
func (f *fakeEngine) Link(ctx context.Context, req *mbp.LinkRequest) (*mbp.LinkResponse, error) {
	return &mbp.LinkResponse{}, nil
}
func (f *fakeEngine) Stat(ctx context.Context, req *mbp.StatRequest) (*mbp.StatResponse, error) {
	return &mbp.StatResponse{}, nil
}
func (f *fakeEngine) GetContradictions(ctx context.Context, vault string) ([]ContradictionPair, error) {
	return nil, nil
}
func (f *fakeEngine) Evolve(ctx context.Context, vault, oldID, newContent, reason string) (*WriteResult, error) {
	return &WriteResult{ID: "new-id"}, nil
}
func (f *fakeEngine) Consolidate(ctx context.Context, vault string, ids []string, merged string) (*ConsolidateResult, error) {
	return &ConsolidateResult{ID: "merged-id"}, nil
}
func (f *fakeEngine) Session(ctx context.Context, vault string, since time.Time) (*SessionSummary, error) {
	return &SessionSummary{}, nil
}
func (f *fakeEngine) Decide(ctx context.Context, vault, decision, rationale string, alternatives, evidenceIDs []string) (*WriteResult, error) {
	return &WriteResult{ID: "decision-id"}, nil
}

// Epic 18 fake implementations
func (f *fakeEngine) Restore(ctx context.Context, vault string, id string) (*RestoreResult, error) {
	return &RestoreResult{ID: id, Concept: "restored concept", State: "active"}, nil
}
func (f *fakeEngine) Traverse(ctx context.Context, vault string, req *TraverseRequest) (*TraverseResult, error) {
	return &TraverseResult{Nodes: []TraversalNode{}, Edges: []TraversalEdge{}}, nil
}
func (f *fakeEngine) Explain(ctx context.Context, vault string, req *ExplainRequest) (*ExplainResult, error) {
	return &ExplainResult{EngramID: req.EngramID, WouldReturn: true, Threshold: 0.5}, nil
}
func (f *fakeEngine) UpdateState(ctx context.Context, vault string, id string, state string, reason string) error {
	return nil
}
func (f *fakeEngine) ListDeleted(ctx context.Context, vault string, limit int) ([]DeletedEngram, error) {
	return []DeletedEngram{}, nil
}
func (f *fakeEngine) RetryEnrich(ctx context.Context, vault string, id string) (*RetryEnrichResult, error) {
	return &RetryEnrichResult{EngramID: id, PluginsQueued: []string{}, AlreadyComplete: []string{}}, nil
}

func newTestServer() *MCPServer {
	return New(":0", &fakeEngine{}, "")
}

func postRPC(t *testing.T, srv *MCPServer, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(w, req)
	return w
}

func TestMalformedJSONRPC(t *testing.T) {
	srv := newTestServer()

	cases := []struct {
		name        string
		body        string
		wantErrCode int
	}{
		{
			name:        "invalid JSON",
			body:        `{not json`,
			wantErrCode: -32700,
		},
		{
			name:        "jsonrpc not 2.0",
			body:        `{"jsonrpc":"1.0","method":"tools/call","id":1}`,
			wantErrCode: -32600,
		},
		{
			name:        "missing method",
			body:        `{"jsonrpc":"2.0","id":1}`,
			wantErrCode: -32601,
		},
		{
			name:        "unknown method",
			body:        `{"jsonrpc":"2.0","method":"unknown","id":1}`,
			wantErrCode: -32601,
		},
		{
			name:        "unknown tool name",
			body:        `{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"muninn_nonexistent","arguments":{"vault":"default"}}}`,
			wantErrCode: -32602,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := postRPC(t, srv, tc.body)
			var resp JSONRPCResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if resp.Error == nil {
				t.Fatal("expected error in response, got nil")
			}
			if resp.Error.Code != tc.wantErrCode {
				t.Errorf("error code = %d, want %d", resp.Error.Code, tc.wantErrCode)
			}
		})
	}
}

func TestBodySizeLimit(t *testing.T) {
	srv := newTestServer()
	big := bytes.Repeat([]byte("x"), 2<<20)
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(big))
	w := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected non-200 for oversized body")
	}
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/mcp/health", nil)
	w := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("health status = %d, want 200", w.Code)
	}
}

func TestListTools(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/mcp/tools", nil)
	w := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("tools status = %d, want 200", w.Code)
	}
	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	tools, _ := result["tools"].([]any)
	if len(tools) != 17 {
		t.Errorf("expected 17 tools, got %d", len(tools))
	}
}

func TestHandleRememberHappyPath(t *testing.T) {
	srv := newTestServer()
	body := `{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"muninn_remember","arguments":{"vault":"default","content":"test content"}}}`
	w := postRPC(t, srv, body)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp JSONRPCResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Error("expected non-nil result")
	}
}

func TestHandleRecallHappyPath(t *testing.T) {
	srv := newTestServer()
	body := `{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"muninn_recall","arguments":{"vault":"default","context":["test query"]}}}`
	w := postRPC(t, srv, body)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp JSONRPCResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestHandleRememberMissingContent(t *testing.T) {
	srv := newTestServer()
	body := `{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"muninn_remember","arguments":{"vault":"default"}}}`
	w := postRPC(t, srv, body)
	var resp JSONRPCResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32602 {
		t.Errorf("expected -32602, got %v", resp.Error)
	}
}

func TestHandleConsolidateExceedsLimit(t *testing.T) {
	srv := newTestServer()
	ids := make([]string, 51)
	for i := range ids {
		ids[i] = "id"
	}
	b, _ := json.Marshal(map[string]any{
		"vault": "default", "ids": ids, "merged_content": "merged",
	})
	body := `{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"muninn_consolidate","arguments":` + string(b) + `}}`
	w := postRPC(t, srv, body)
	var resp JSONRPCResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32602 {
		t.Errorf("expected -32602 for >50 ids, got %v", resp.Error)
	}
}

func TestHandleEvolveRequiredParams(t *testing.T) {
	srv := newTestServer()
	// Missing new_content
	body := `{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"muninn_evolve","arguments":{"vault":"default","id":"abc","reason":"why"}}}`
	w := postRPC(t, srv, body)
	var resp JSONRPCResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32602 {
		t.Errorf("expected -32602 for missing new_content, got %v", resp.Error)
	}
}

func TestHandleSessionInvalidSince(t *testing.T) {
	srv := newTestServer()
	body := `{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"muninn_session","arguments":{"vault":"default","since":"not-a-timestamp"}}}`
	w := postRPC(t, srv, body)
	var resp JSONRPCResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32602 {
		t.Errorf("expected -32602 for invalid since, got %v", resp.Error)
	}
}
