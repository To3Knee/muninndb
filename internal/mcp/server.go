package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// MCPServer serves the MCP JSON-RPC 2.0 protocol on a single HTTP mux.
type MCPServer struct {
	engine  EngineInterface
	token   string // required Bearer token; empty = no auth
	limiter *rate.Limiter
	srv     *http.Server
}

// New creates an MCPServer. addr is the listen address (e.g., ":8750").
// token is the required Bearer token; pass "" to disable auth.
func New(addr string, eng EngineInterface, token string) *MCPServer {
	s := &MCPServer{
		engine:  eng,
		token:   token,
		limiter: rate.NewLimiter(rate.Limit(100), 200),
	}
	mux := http.NewServeMux()
	// Use method-agnostic paths with manual method dispatch for compatibility
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.withMiddleware(s.handleRPC)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/mcp/tools", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.withMiddleware(s.handleListTools)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/mcp/health", s.handleHealth)
	s.srv = &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	return s
}

// Serve starts listening. Blocks until the server is stopped.
func (s *MCPServer) Serve() error { return s.srv.ListenAndServe() }

// Shutdown gracefully stops the server.
func (s *MCPServer) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }

// withMiddleware wraps a handler with: body size limit → rate limiter → auth check.
func (s *MCPServer) withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enforce 1MB body limit before any processing.
		if r.ContentLength > 1<<20 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32700,"message":"request body too large"}}`))
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		if !s.limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32000,"message":"rate limited"}}`))
			return
		}
		auth := authFromRequest(r, s.token)
		if !auth.Authorized {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32001,"message":"unauthorized"}}`))
			return
		}
		next(w, r)
	}
}

func (s *MCPServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, nil, -32700, "parse error")
		return
	}
	if req.JSONRPC != "2.0" {
		sendError(w, req.ID, -32600, "invalid request: jsonrpc must be '2.0'")
		return
	}

	switch req.Method {
	case "tools/list":
		sendResult(w, req.ID, map[string]any{"tools": allToolDefinitions()})
	case "tools/call":
		s.dispatchToolCall(ctx, w, &req)
	case "":
		sendError(w, req.ID, -32601, "method not found: method is required")
	default:
		sendError(w, req.ID, -32601, "method not found: "+req.Method)
	}
}

func (s *MCPServer) dispatchToolCall(ctx context.Context, w http.ResponseWriter, req *JSONRPCRequest) {
	if req.Params == nil {
		sendError(w, req.ID, -32602, "invalid params: params required")
		return
	}

	args := req.Params.Arguments
	if args == nil {
		args = make(map[string]any)
	}

	vault, ok := vaultFromArgs(args)
	if !ok {
		sendError(w, req.ID, -32602, "invalid params: 'vault' is required (use 'default' for the default vault)")
		return
	}

	handlers := map[string]func(context.Context, http.ResponseWriter, json.RawMessage, string, map[string]any){
		"muninn_remember":       s.handleRemember,
		"muninn_recall":         s.handleRecall,
		"muninn_read":           s.handleRead,
		"muninn_forget":         s.handleForget,
		"muninn_link":           s.handleLink,
		"muninn_contradictions": s.handleContradictions,
		"muninn_status":         s.handleStatus,
		"muninn_evolve":         s.handleEvolve,
		"muninn_consolidate":    s.handleConsolidate,
		"muninn_session":        s.handleSession,
		"muninn_decide":         s.handleDecide,
		// Epic 18: tools 12-17
		"muninn_restore":      s.handleRestore,
		"muninn_traverse":     s.handleTraverse,
		"muninn_explain":      s.handleExplain,
		"muninn_state":        s.handleState,
		"muninn_list_deleted": s.handleListDeleted,
		"muninn_retry_enrich": s.handleRetryEnrich,
	}

	handler, found := handlers[req.Params.Name]
	if !found {
		sendError(w, req.ID, -32602, "unknown tool: "+req.Params.Name)
		return
	}
	handler(ctx, w, req.ID, vault, args)
}

func (s *MCPServer) handleListTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"tools": allToolDefinitions()})
}

func (s *MCPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// sendResult writes a successful JSON-RPC response.
func sendResult(w http.ResponseWriter, id json.RawMessage, result any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

// sendError writes a JSON-RPC error response.
func sendError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &JSONRPCError{Code: code, Message: message},
	})
}

// mustJSON marshals v to JSON.
// On marshal failure it logs the error and returns an empty JSON object
// rather than panicking — marshal errors are caused by non-serialisable types
// in dynamic handler data, not programmer bugs in static schema definitions.
func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		slog.Error("mcp: mustJSON marshal failed", "error", err)
		return "{}"
	}
	return string(b)
}
