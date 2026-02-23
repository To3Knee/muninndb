# MCP Sessions, Vault Pinning & SSE Push — Design

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Bring the MCP server to spec compliance (initialize handshake), enable per-project vault pinning via URL query param, and add SSE server-push for cognitive events (contradictions, activations, Hebbian associations).

**Architecture:** Sessions are established via the MCP `initialize` handshake. The vault is pinned per-session from the `?vault=` URL param. A `GET /mcp` SSE endpoint allows the server to push cognitive notifications to connected AI tools. An `EventBus` bridges the engine's trigger/worker system to active sessions.

**Tech Stack:** Go, `sync`, `atomic`, `crypto/rand`, `encoding/json`, `net/http` (SSE via `http.Flusher`)

---

## Background

### What we fixed
The MCP server (`internal/mcp/`) was stateless HTTP JSON-RPC. It did not handle the `initialize` method required by the MCP spec (v2025-03-26). Clients (Claude Desktop, Claude Code, Cursor) were tolerant of this but it was non-compliant. Additionally, every tool call required `vault` as a required argument, with no way to pin a session to a specific vault.

### Auth model
- Global bearer token: proves "you are the machine owner." Checked on every request.
- Session token binding: `sha256(bearerToken)` stored in session at creation, verified on every subsequent request. Prevents stale sessions from being used after token rotation.
- `Mcp-Session-Id`: 256-bit `crypto/rand` hex string. Routing mechanism, not standalone auth.

### Vault routing (final resolution order)
1. Session-pinned vault (from `?vault=` at `initialize` time) — if session exists and arg matches or is absent → use session vault; if arg differs → error `-32003 vault mismatch`
2. No session + explicit `vault` arg → use arg
3. No session + no arg → `"default"`

---

## Session Design

### `mcpSession` struct
```go
type mcpSession struct {
    vault         string
    tokenHash     [32]byte       // sha256(bearerToken) at creation
    createdAt     time.Time
    lastUsed      time.Time
    initialized   bool           // true after "initialized" notification
    streaming     atomic.Bool    // true while SSE client connected
    droppedEvents atomic.Int64   // events dropped due to full pushCh
    pushCh        chan MCPEvent   // buffered (default 64); closed only by session reaper
}
```

### `SessionStore` interface
```go
type SessionStore interface {
    Create(vault string, tokenHash [32]byte) (sessionID string, err error)
    Get(sessionID string) (*mcpSession, bool)
    Touch(sessionID string)
    MarkInitialized(sessionID string) error
    ByVault(vault string) []*mcpSession  // returns snapshot copy (not internal reference)
    DroppedCount(sessionID string) int64
    Close()
}
```

Concrete implementation:
- `sync.RWMutex` + `map[string]*mcpSession`
- Cap: 256 sessions; `Create()` returns error if at cap
- TTL: 24h; cleanup goroutine uses `done chan struct{}` for clean shutdown
- Push channel buffer: 64 (configurable via constructor param; absorbs consolidation bursts)
- Injectable `now func() time.Time` for test clock control
- **Only the cleanup goroutine closes `pushCh`** — after setting `session.streaming.Store(false)`

---

## EventBus Design

```go
type MCPEvent struct {
    Vault  string          `json:"vault"`
    Method string          `json:"method"`  // e.g. "notifications/muninn/contradiction"
    Params json.RawMessage `json:"params"`  // pre-serialized at emission site
}

type EventBus interface {
    Publish(event MCPEvent)
    Events() <-chan MCPEvent  // single subscriber: MCPServer only
    Close()
}
```

Concrete implementation:
- Buffered channel, size 256
- `Publish()` non-blocking: drop + increment global bus-level drop counter if full
- `Close()` closes the channel, terminating the MCPServer's fan-out goroutine
- Injected into engine as optional interface: `nil` = no-op (push disabled)

### Cognitive event types
Pre-serialize `Params` at emission site so marshaling errors surface at the source:

```go
// Contradiction detected
params, _ := json.Marshal(struct{ A, B, Concept string }{idA, idB, concept})
bus.Publish(MCPEvent{Vault: vault, Method: "notifications/muninn/contradiction", Params: params})

// Background activation
params, _ := json.Marshal(struct{ ID, Concept string; Score float32 }{id, concept, score})
bus.Publish(MCPEvent{Vault: vault, Method: "notifications/muninn/activation", Params: params})

// Hebbian association formed
params, _ := json.Marshal(struct{ SourceID, TargetID string; Weight float32 }{src, tgt, w})
bus.Publish(MCPEvent{Vault: vault, Method: "notifications/muninn/association", Params: params})
```

---

## MCP Server Protocol Flow

### initialize
```
POST /mcp?vault=my-project
Body: {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"Cursor","version":"1.0"}}}

Response header: Mcp-Session-Id: <256-bit hex>
Response body:   {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-03-26","capabilities":{"tools":{}},"serverInfo":{"name":"muninn","version":"..."}}}
```

### initialized (notification, no response body)
```
POST /mcp
Headers: Mcp-Session-Id: <id>
Body: {"jsonrpc":"2.0","method":"notifications/initialized"}
Response: 202 Accepted
```

Session `initialized` flag set to `true`. Tool calls before this → `-32002`.

### Tool calls (vault optional)
```
POST /mcp
Headers: Authorization: Bearer <token>, Mcp-Session-Id: <id>
Body: {"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"muninn_remember","arguments":{"content":"..."}}}
```
`vault` is now optional in all tool schemas. Resolved via vault resolution order above.

### SSE push
```
GET /mcp
Headers: Authorization: Bearer <token>, Mcp-Session-Id: <id>, Accept: text/event-stream

Response: Content-Type: text/event-stream (keep-alive)
event: notification
data: {"jsonrpc":"2.0","method":"notifications/muninn/contradiction","params":{"a":"...","b":"..."}}

```
- 405 if `Accept: text/event-stream` missing
- 401 if token invalid or tokenHash mismatch
- 400 if session not found or not initialized

---

## Fan-out Goroutine

```go
// MCPServer startup
go func() {
    for event := range s.events.Events() {
        sessions := s.store.ByVault(event.Vault)  // snapshot copy
        for _, sess := range sessions {
            sess := sess
            go func() {
                if !sess.streaming.Load() { return }
                select {
                case sess.pushCh <- event:
                default:
                    sess.droppedEvents.Add(1)
                }
            }()
        }
    }
}()
```

---

## Shutdown Ordering (must be followed exactly)

1. Stop accepting new connections (`srv.Shutdown(ctx)`)
2. Close EventBus → closes `Events()` channel → terminates fan-out goroutine
3. Session reaper: close all `pushCh` channels → terminates SSE goroutines
4. Close SessionStore

Document this ordering in a comment in `MCPServer.Shutdown()`.

---

## Observability

### Admin endpoint: `GET /api/admin/mcp/sessions`
```json
{
  "sessions": [
    {
      "session_id": "a1b2...",
      "vault": "my-project",
      "initialized": true,
      "streaming": true,
      "dropped_events": 0,
      "created_at": "2026-02-22T10:00:00Z",
      "last_used": "2026-02-22T10:05:00Z"
    }
  ],
  "bus_dropped_events": 0,
  "session_count": 1,
  "session_cap": 256
}
```

---

## Files Changed

| File | Change |
|------|--------|
| `internal/mcp/server.go` | Add initialize/initialized handlers, GET SSE endpoint, fan-out goroutine, shutdown ordering |
| `internal/mcp/session.go` | New: SessionStore interface + concrete impl |
| `internal/mcp/eventbus.go` | New: EventBus interface + concrete impl |
| `internal/mcp/context.go` | Vault resolution logic (session-aware) |
| `internal/mcp/types.go` | MCPEvent, cognitive param structs |
| `internal/mcp/tools.go` | Make `vault` optional in all 17 tool schemas |
| `internal/engine/engine.go` | Accept EventBus as optional dep (nil-safe) |
| `internal/engine/trigger/system.go` | Emit contradiction + activation events |
| `internal/cognitive/worker.go` | Emit Hebbian association events |
| `internal/transport/rest/admin_handlers.go` | Add GET /api/admin/mcp/sessions handler |
| `internal/mcp/session_test.go` | New: SessionStore tests (fake clock, interface) |
| `internal/mcp/server_test.go` | Add initialize flow, SSE push, vault mismatch tests |

## What Does NOT Change

- REST API, storage layer, auth store
- Existing tool call behavior with explicit vault and no session
- Bearer token format or global token file location
