# MCP Sessions, Vault Pinning & SSE Push — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Add MCP spec-compliant session initialization, per-project vault pinning via `?vault=` URL param, and SSE server-push for cognitive events (contradictions, activations, Hebbian associations).

**Architecture:** Sessions are created during the `initialize` JSON-RPC handshake. An `EventBus` bridges the engine's trigger/worker system to active sessions. A `GET /mcp` SSE endpoint streams cognitive notifications to connected AI tools. Three-round Opus-approved design — do not deviate from it.

**Tech Stack:** Go stdlib (`sync`, `sync/atomic`, `crypto/rand`, `encoding/json`, `net/http`, `fmt`), no new dependencies.

**Design doc:** `docs/plans/2026-02-22-mcp-sessions-vault-push-design.md` — read it before starting.

**Test command:** `go test ./internal/mcp/... -race -count=1`
**Build verify:** `go build ./...`

---

## Key architectural rules (read before touching any file)

1. **`pushCh` is closed ONLY by the session reaper** (cleanup goroutine). The SSE goroutine sets `streaming.Store(false)` on disconnect but never closes the channel.
2. **`initialized` flag** — sessions start as `false`. Tool calls and SSE reject with `-32002` until the client sends the `initialized` notification.
3. **Sessions only created in `initialize` handler** — never in tool calls or SSE.
4. **Vault resolution order:** session-pinned → explicit arg (no session) → `"default"`. Session-pinned + differing arg = `-32003` error.
5. **Fan-out is non-blocking:** drop events when `pushCh` full, increment counter. Engine must never block on AI client backpressure.
6. **Shutdown ordering:** stop accepting → close EventBus → reaper closes pushChs → close SessionStore.
7. **Bearer token validated on every request** including GET SSE. `sha256(token)` verified against `session.tokenHash`.
8. **`Params` in MCPEvent is `json.RawMessage`** — pre-serialized at emission site.

---

## Task 1: EventBus — interface + implementation

**Files:**
- Create: `internal/mcp/eventbus.go`
- Create: `internal/mcp/eventbus_test.go`

**Step 1: Write failing tests**

```go
// internal/mcp/eventbus_test.go
package mcp

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventBus_PublishAndReceive(t *testing.T) {
	bus := NewEventBus(8)
	defer bus.Close()

	params, _ := json.Marshal(map[string]string{"a": "1"})
	evt := MCPEvent{Vault: "test", Method: "notifications/muninn/test", Params: params}
	bus.Publish(evt)

	select {
	case got := <-bus.Events():
		if got.Vault != "test" || got.Method != "notifications/muninn/test" {
			t.Fatalf("unexpected event: %+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBus_DropsWhenFull(t *testing.T) {
	bus := NewEventBus(1)
	defer bus.Close()

	params, _ := json.Marshal(map[string]string{})
	evt := MCPEvent{Vault: "v", Method: "m", Params: params}

	bus.Publish(evt) // fills buffer
	bus.Publish(evt) // should drop

	if bus.BusDropped() != 1 {
		t.Fatalf("expected 1 dropped, got %d", bus.BusDropped())
	}
}

func TestEventBus_CloseTerminatesEvents(t *testing.T) {
	bus := NewEventBus(8)
	bus.Close()
	// Events() channel should be closed (range terminates)
	count := 0
	for range bus.Events() {
		count++
	}
	if count != 0 {
		t.Fatalf("expected 0 events after close, got %d", count)
	}
}
```

**Step 2: Run to verify failure**

```bash
go test ./internal/mcp/... -run TestEventBus -v
```
Expected: `FAIL` — `MCPEvent`, `NewEventBus`, `BusDropped` undefined.

**Step 3: Implement**

```go
// internal/mcp/eventbus.go
package mcp

import (
	"encoding/json"
	"sync/atomic"
)

// MCPEvent is a cognitive notification sent from the engine to connected AI tools.
// Params must be pre-serialized JSON — marshaling errors are caught at the emission site.
type MCPEvent struct {
	Vault  string          `json:"vault"`
	Method string          `json:"method"` // e.g. "notifications/muninn/contradiction"
	Params json.RawMessage `json:"params"`
}

// EventBus delivers cognitive events from the engine to the MCP server.
// There is exactly one subscriber: MCPServer. Events() returns the same channel on every call.
type EventBus interface {
	Publish(event MCPEvent)
	Events() <-chan MCPEvent
	BusDropped() int64 // events dropped at bus level (buffer full)
	Close()
}

// NoopEventBus is a nil-safe EventBus used when push is disabled.
type NoopEventBus struct{}

func (NoopEventBus) Publish(MCPEvent)        {}
func (NoopEventBus) Events() <-chan MCPEvent { return nil }
func (NoopEventBus) BusDropped() int64       { return 0 }
func (NoopEventBus) Close()                  {}

type chanEventBus struct {
	ch      chan MCPEvent
	dropped atomic.Int64
}

// NewEventBus creates an EventBus with the given channel buffer size.
// bufSize of 256 is recommended for production.
func NewEventBus(bufSize int) EventBus {
	return &chanEventBus{ch: make(chan MCPEvent, bufSize)}
}

func (b *chanEventBus) Publish(event MCPEvent) {
	select {
	case b.ch <- event:
	default:
		b.dropped.Add(1)
	}
}

func (b *chanEventBus) Events() <-chan MCPEvent { return b.ch }
func (b *chanEventBus) BusDropped() int64       { return b.dropped.Load() }
func (b *chanEventBus) Close()                  { close(b.ch) }
```

**Step 4: Run tests**

```bash
go test ./internal/mcp/... -run TestEventBus -v -race
```
Expected: all pass.

**Step 5: Commit**

```bash
git add internal/mcp/eventbus.go internal/mcp/eventbus_test.go
git commit -m "feat(mcp): EventBus interface and implementation"
```

---

## Task 2: MCPEvent cognitive param types

**Files:**
- Modify: `internal/mcp/types.go` (append at end)

These are the param structs pre-serialized at emission sites (trigger system, cognitive worker).

**Step 1: Append to `internal/mcp/types.go`**

```go
// --- Cognitive push notification param types ---
// These are pre-serialized to json.RawMessage at emission sites.

// ContradictionParams is the params payload for "notifications/muninn/contradiction".
type ContradictionParams struct {
	IDa     string `json:"id_a"`
	IDb     string `json:"id_b"`
	Concept string `json:"concept,omitempty"`
}

// ActivationParams is the params payload for "notifications/muninn/activation".
type ActivationParams struct {
	ID      string  `json:"id"`
	Concept string  `json:"concept"`
	Score   float64 `json:"score"`
	Vault   string  `json:"vault"`
}

// AssociationParams is the params payload for "notifications/muninn/association".
type AssociationParams struct {
	SourceID string  `json:"source_id"`
	TargetID string  `json:"target_id"`
	Weight   float32 `json:"weight"`
}
```

**Step 2: Build verify**

```bash
go build ./internal/mcp/...
```
Expected: no errors.

**Step 3: Commit**

```bash
git add internal/mcp/types.go
git commit -m "feat(mcp): cognitive push notification param types"
```

---

## Task 3: SessionStore — interface + implementation

**Files:**
- Create: `internal/mcp/session.go`
- Create: `internal/mcp/session_test.go`

**Step 1: Write failing tests**

```go
// internal/mcp/session_test.go
package mcp

import (
	"crypto/sha256"
	"testing"
	"time"
)

func TestSessionStore_CreateAndGet(t *testing.T) {
	now := time.Now()
	store := newSessionStore(8, func() time.Time { return now })
	defer store.Close()

	tokenHash := sha256.Sum256([]byte("mytoken"))
	id, err := store.Create("myvault", tokenHash)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id == "" {
		t.Fatal("empty session ID")
	}

	sess, ok := store.Get(id)
	if !ok {
		t.Fatal("expected session to exist")
	}
	if sess.vault != "myvault" {
		t.Fatalf("vault: got %q, want %q", sess.vault, "myvault")
	}
	if sess.tokenHash != tokenHash {
		t.Fatal("tokenHash mismatch")
	}
	if sess.initialized.Load() {
		t.Fatal("session should not be initialized yet")
	}
}

func TestSessionStore_MarkInitialized(t *testing.T) {
	store := newSessionStore(8, time.Now)
	defer store.Close()

	tokenHash := sha256.Sum256([]byte("t"))
	id, _ := store.Create("v", tokenHash)

	if err := store.MarkInitialized(id); err != nil {
		t.Fatalf("MarkInitialized: %v", err)
	}
	sess, _ := store.Get(id)
	if !sess.initialized.Load() {
		t.Fatal("expected initialized == true")
	}
}

func TestSessionStore_Cap(t *testing.T) {
	store := newSessionStore(2, time.Now)
	defer store.Close()

	h := sha256.Sum256([]byte("t"))
	store.Create("v", h)
	store.Create("v", h)
	_, err := store.Create("v", h) // should fail
	if err == nil {
		t.Fatal("expected cap error, got nil")
	}
}

func TestSessionStore_TTLExpiry(t *testing.T) {
	tick := time.Now()
	store := newSessionStore(8, func() time.Time { return tick })
	defer store.Close()

	h := sha256.Sum256([]byte("t"))
	id, _ := store.Create("v", h)

	// advance clock past TTL
	tick = tick.Add(25 * time.Hour)
	store.(*concreteSessionStore).sweep()

	_, ok := store.Get(id)
	if ok {
		t.Fatal("expected session to be expired and removed")
	}
}

func TestSessionStore_ByVault(t *testing.T) {
	store := newSessionStore(8, time.Now)
	defer store.Close()

	h := sha256.Sum256([]byte("t"))
	store.Create("vault-a", h)
	store.Create("vault-a", h)
	store.Create("vault-b", h)

	sessions := store.ByVault("vault-a")
	if len(sessions) != 2 {
		t.Fatalf("ByVault: got %d, want 2", len(sessions))
	}
}

func TestSessionStore_VaultMismatch(t *testing.T) {
	// vault mismatch detection is in vault resolution, not store — just verify
	// that Get returns the session with its vault for the caller to check.
	store := newSessionStore(8, time.Now)
	defer store.Close()

	h := sha256.Sum256([]byte("t"))
	id, _ := store.Create("project-a", h)
	sess, ok := store.Get(id)
	if !ok || sess.vault != "project-a" {
		t.Fatal("unexpected vault in session")
	}
}
```

**Step 2: Run to verify failure**

```bash
go test ./internal/mcp/... -run TestSessionStore -v
```
Expected: `FAIL` — types undefined.

**Step 3: Implement**

```go
// internal/mcp/session.go
package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	sessionTTL        = 24 * time.Hour
	defaultPushBufSz  = 64 // absorbs consolidation bursts; drops acceptable beyond this
)

// ErrSessionCapReached is returned by Create when the session cap is hit.
var ErrSessionCapReached = errors.New("mcp: session cap reached, too many active sessions")

// mcpSession holds per-session state. Fields accessed concurrently use atomics.
// pushCh is closed ONLY by the session reaper — never by the SSE goroutine.
type mcpSession struct {
	vault         string
	tokenHash     [32]byte
	createdAt     time.Time
	lastUsed      time.Time // protected by store.mu
	initialized   atomic.Bool
	streaming     atomic.Bool
	droppedEvents atomic.Int64
	pushCh        chan MCPEvent
}

// SessionStore manages active MCP sessions.
type SessionStore interface {
	Create(vault string, tokenHash [32]byte) (sessionID string, err error)
	Get(sessionID string) (*mcpSession, bool)
	Touch(sessionID string)
	MarkInitialized(sessionID string) error
	ByVault(vault string) []*mcpSession // returns snapshot — safe to iterate after lock released
	DroppedCount(sessionID string) int64
	Close()
}

type concreteSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*mcpSession
	cap      int
	now      func() time.Time
	pushBuf  int
	done     chan struct{}
}

func newSessionStore(cap int, now func() time.Time) SessionStore {
	s := &concreteSessionStore{
		sessions: make(map[string]*mcpSession),
		cap:      cap,
		now:      now,
		pushBuf:  defaultPushBufSz,
		done:     make(chan struct{}),
	}
	go s.cleanupLoop()
	return s
}

func (s *concreteSessionStore) Create(vault string, tokenHash [32]byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.sessions) >= s.cap {
		return "", fmt.Errorf("%w (cap=%d)", ErrSessionCapReached, s.cap)
	}
	id, err := newSessionID()
	if err != nil {
		return "", err
	}
	sess := &mcpSession{
		vault:     vault,
		tokenHash: tokenHash,
		createdAt: s.now(),
		lastUsed:  s.now(),
		pushCh:    make(chan MCPEvent, s.pushBuf),
	}
	s.sessions[id] = sess
	return id, nil
}

func (s *concreteSessionStore) Get(sessionID string) (*mcpSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	return sess, ok
}

func (s *concreteSessionStore) Touch(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[sessionID]; ok {
		sess.lastUsed = s.now()
	}
}

func (s *concreteSessionStore) MarkInitialized(sessionID string) error {
	s.mu.RLock()
	sess, ok := s.sessions[sessionID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	sess.initialized.Store(true)
	return nil
}

func (s *concreteSessionStore) ByVault(vault string) []*mcpSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*mcpSession
	for _, sess := range s.sessions {
		if sess.vault == vault {
			result = append(result, sess)
		}
	}
	return result // snapshot slice; elements are pointers, atomic fields are safe
}

func (s *concreteSessionStore) DroppedCount(sessionID string) int64 {
	s.mu.RLock()
	sess, ok := s.sessions[sessionID]
	s.mu.RUnlock()
	if !ok {
		return 0
	}
	return sess.droppedEvents.Load()
}

// Close stops the cleanup goroutine. Called as part of server shutdown.
// IMPORTANT: the shutdown sequence in MCPServer.Shutdown() calls Close()
// which sweeps and closes all pushCh channels, terminating SSE goroutines.
func (s *concreteSessionStore) Close() {
	close(s.done)
	s.sweep() // final sweep: close all pushCh channels
}

func (s *concreteSessionStore) cleanupLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.sweep()
		case <-s.done:
			return
		}
	}
}

// sweep removes expired sessions and closes their pushCh channels.
// This is the ONLY place that closes pushCh.
func (s *concreteSessionStore) sweep() {
	s.mu.Lock()
	now := s.now()
	var expired []*mcpSession
	for id, sess := range s.sessions {
		if now.Sub(sess.lastUsed) > sessionTTL {
			expired = append(expired, sess)
			delete(s.sessions, id)
		}
	}
	s.mu.Unlock()

	// Close channels outside the lock — SSE goroutines will unblock
	for _, sess := range expired {
		sess.streaming.Store(false)
		close(sess.pushCh)
	}
}

// newSessionID generates a 256-bit cryptographically random session ID.
func newSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate session ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Compile-time assertion that concreteSessionStore implements SessionStore.
var _ SessionStore = (*concreteSessionStore)(nil)
```

**Step 4: Run tests**

```bash
go test ./internal/mcp/... -run TestSessionStore -v -race
```
Expected: all pass.

**Step 5: Commit**

```bash
git add internal/mcp/session.go internal/mcp/session_test.go
git commit -m "feat(mcp): SessionStore interface and implementation"
```

---

## Task 4: Vault resolution — session-aware context

**Files:**
- Modify: `internal/mcp/context.go`

Replace the current file entirely. The current `authFromRequest` and `vaultFromArgs` are preserved; new session-aware vault resolution is added.

**Step 1: Write the new context.go**

```go
// internal/mcp/context.go
package mcp

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
)

const mcpSessionHeader = "Mcp-Session-Id"

// authFromRequest extracts the Bearer token from the Authorization header.
// Returns AuthContext{Authorized: true} if token matches or no token is required.
func authFromRequest(r *http.Request, requiredToken string) AuthContext {
	if requiredToken == "" {
		return AuthContext{Authorized: true}
	}
	header := r.Header.Get("Authorization")
	token, found := strings.CutPrefix(header, "Bearer ")
	if !found || token == "" {
		return AuthContext{Authorized: false}
	}
	return AuthContext{Token: token, Authorized: token == requiredToken}
}

// sessionFromRequest looks up a session by the Mcp-Session-Id header.
// Returns (nil, "") if no header present.
// Returns (nil, sessionID) if header present but session not found or expired.
func sessionFromRequest(r *http.Request, store SessionStore) (sess *mcpSession, sessionID string) {
	sessionID = r.Header.Get(mcpSessionHeader)
	if sessionID == "" {
		return nil, ""
	}
	sess, ok := store.Get(sessionID)
	if !ok {
		return nil, sessionID
	}
	return sess, sessionID
}

// validateSessionToken checks that the bearer token matches the session's token hash.
// Returns an error string if invalid, "" if valid.
func validateSessionToken(sess *mcpSession, token string) string {
	h := sha256.Sum256([]byte(token))
	if h != sess.tokenHash {
		return "token does not match session"
	}
	return ""
}

// resolveVault determines the effective vault for a tool call.
//
// Resolution order (per Opus-approved design):
//  1. Session pinned vault — if session exists and arg matches or is absent: use session vault
//  2. Session pinned vault — if arg differs: return vault mismatch error
//  3. No session + explicit arg: use arg
//  4. No session + no arg: use "default"
//
// Returns (vault, errMsg). errMsg is non-empty on error.
func resolveVault(sess *mcpSession, args map[string]any) (vault string, errMsg string) {
	argVault, hasArg := vaultFromArgs(args)

	if sess != nil {
		if !hasArg || argVault == "" || argVault == sess.vault {
			return sess.vault, ""
		}
		// Arg present and differs from session pin — this is almost always a bug
		return "", fmt.Sprintf(
			"vault mismatch: session pinned to %q but tool call specified %q — "+
				"omit vault arg or match the session vault",
			sess.vault, argVault,
		)
	}

	if hasArg && argVault != "" {
		return argVault, ""
	}
	return "default", ""
}

// vaultFromArgs extracts the vault parameter from tool arguments.
// Returns ("", false) if vault is missing or empty.
func vaultFromArgs(args map[string]any) (string, bool) {
	v, ok := args["vault"]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", false
	}
	return s, true
}
```

**Step 2: Write tests for vault resolution**

Add to `internal/mcp/session_test.go`:

```go
func TestResolveVault_SessionPin(t *testing.T) {
	sess := &mcpSession{vault: "project-a"}
	vault, errMsg := resolveVault(sess, map[string]any{})
	if vault != "project-a" || errMsg != "" {
		t.Fatalf("got vault=%q err=%q", vault, errMsg)
	}
}

func TestResolveVault_Mismatch(t *testing.T) {
	sess := &mcpSession{vault: "project-a"}
	_, errMsg := resolveVault(sess, map[string]any{"vault": "project-b"})
	if errMsg == "" {
		t.Fatal("expected vault mismatch error")
	}
}

func TestResolveVault_NoSessionWithArg(t *testing.T) {
	vault, errMsg := resolveVault(nil, map[string]any{"vault": "explicit"})
	if vault != "explicit" || errMsg != "" {
		t.Fatalf("got vault=%q err=%q", vault, errMsg)
	}
}

func TestResolveVault_DefaultFallback(t *testing.T) {
	vault, errMsg := resolveVault(nil, map[string]any{})
	if vault != "default" || errMsg != "" {
		t.Fatalf("got vault=%q err=%q", vault, errMsg)
	}
}
```

**Step 3: Run tests**

```bash
go test ./internal/mcp/... -run "TestResolveVault|TestSessionStore" -v -race
```
Expected: all pass.

**Step 4: Commit**

```bash
git add internal/mcp/context.go internal/mcp/session_test.go
git commit -m "feat(mcp): session-aware vault resolution"
```

---

## Task 5: Make `vault` optional in all 17 tool schemas

**Files:**
- Modify: `internal/mcp/tools.go`

The only change is removing `"vault"` from every `"required"` array. The `vault` property itself stays in the schema (for backward compat with clients that still send it).

**Step 1: In `allToolDefinitions()`, for each tool, change `"required": []string{"vault", ...}` to `"required": []string{...}` (removing `"vault"` from the list)**

For tools with only `vault` required (e.g. `muninn_contradictions`, `muninn_status`), the required array becomes empty: `"required": []string{}`.

Here is the full replacement for the `required` lines of each tool:

```
muninn_remember:      "required": []string{"content"}
muninn_recall:        "required": []string{"context"}
muninn_read:          "required": []string{"id"}
muninn_forget:        "required": []string{"id"}
muninn_link:          "required": []string{"source_id", "target_id", "relation"}
muninn_contradictions:"required": []string{}
muninn_status:        "required": []string{}
muninn_evolve:        "required": []string{"id", "new_content", "reason"}
muninn_consolidate:   "required": []string{"ids", "merged_content"}
muninn_session:       "required": []string{"since"}
muninn_decide:        "required": []string{"decision", "rationale"}
muninn_restore:       "required": []string{"id"}
muninn_traverse:      "required": []string{"start_id"}
muninn_explain:       "required": []string{"engram_id", "query"}
muninn_state:         "required": []string{"id", "state"}
muninn_list_deleted:  "required": []string{}
muninn_retry_enrich:  "required": []string{"id"}
```

Also update `vault` property description for all tools to:
```go
vaultProp := map[string]any{
    "type":        "string",
    "description": "Vault name to scope the operation. Optional when connected via a vault-pinned MCP session. Use 'default' for the default vault.",
}
```

**Step 2: Build verify**

```bash
go build ./internal/mcp/...
```
Expected: no errors.

**Step 3: Run existing MCP tests**

```bash
go test ./internal/mcp/... -v -race
```
Expected: all pass (tool schema tests should still pass — schemas are valid JSON).

**Step 4: Commit**

```bash
git add internal/mcp/tools.go
git commit -m "feat(mcp): make vault optional in all tool schemas"
```

---

## Task 6: MCP server — `initialize` and `initialized` handlers

**Files:**
- Modify: `internal/mcp/server.go`

This is the largest change. Add session store and event bus to `MCPServer`, add the two new handlers, update `dispatchToolCall` to use session-aware vault resolution, and update `withMiddleware` for session validation.

**Step 1: Update `MCPServer` struct and `New()` function**

Replace the struct and constructor:

```go
// MCPServer serves the MCP JSON-RPC 2.0 protocol on a single HTTP mux.
type MCPServer struct {
	engine  EngineInterface
	token   string // required Bearer token; empty = no auth
	limiter *rate.Limiter
	srv     *http.Server
	store   SessionStore
	bus     EventBus
	fanDone chan struct{} // closed when fan-out goroutine exits
}

// New creates an MCPServer. addr is the listen address (e.g., ":8750").
// token is the required Bearer token; pass "" to disable auth.
// bus may be NoopEventBus{} to disable push.
func New(addr string, eng EngineInterface, token string, bus EventBus) *MCPServer {
	if bus == nil {
		bus = NoopEventBus{}
	}
	s := &MCPServer{
		engine:  eng,
		token:   token,
		limiter: rate.NewLimiter(rate.Limit(100), 200),
		store:   newSessionStore(256, time.Now),
		bus:     bus,
		fanDone: make(chan struct{}),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.withMiddleware(s.handleRPC)(w, r)
		case http.MethodGet:
			s.handleSSE(w, r)
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
	go s.fanOutLoop()
	return s
}
```

**Step 2: Update `Shutdown()` with correct ordering**

```go
// Shutdown gracefully stops the server.
// Shutdown ordering (must be preserved):
//  1. Stop accepting new connections (srv.Shutdown)
//  2. Close EventBus — terminates fan-out goroutine
//  3. Session store Close — reaper sweeps, closes all pushCh channels, terminates SSE goroutines
func (s *MCPServer) Shutdown(ctx context.Context) error {
	err := s.srv.Shutdown(ctx)
	s.bus.Close()
	<-s.fanDone // wait for fan-out goroutine to exit
	s.store.Close()
	return err
}
```

**Step 3: Add `handleRPC` changes — add `initialize` and `initialized` cases**

In the `switch req.Method` block, add before `case ""`:

```go
case "initialize":
    s.handleInitialize(w, r, &req)
    return
case "notifications/initialized":
    s.handleInitialized(w, r, &req)
    return
case "ping":
    sendResult(w, req.ID, map[string]any{})
    return
```

**Step 4: Add `handleInitialize`**

```go
func (s *MCPServer) handleInitialize(w http.ResponseWriter, r *http.Request, req *JSONRPCRequest) {
	// Auth — bearer token required even for initialize
	auth := authFromRequest(r, s.token)
	if !auth.Authorized {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32001,"message":"unauthorized"}}`))
		return
	}

	// Vault from URL query param — pinned for this session
	vault := r.URL.Query().Get("vault")

	var tokenHash [32]byte
	if s.token != "" {
		tokenHash = sha256.Sum256([]byte(auth.Token))
	}

	sessionID, err := s.store.Create(vault, tokenHash)
	if err != nil {
		sendError(w, req.ID, -32000, "server error: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(mcpSessionHeader, sessionID)
	sendResult(w, req.ID, map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      map[string]any{"name": "muninn", "version": "1.0"},
	})
}
```

Add import: `"crypto/sha256"` at top of server.go.

**Step 5: Add `handleInitialized`**

```go
func (s *MCPServer) handleInitialized(w http.ResponseWriter, r *http.Request, req *JSONRPCRequest) {
	sessionID := r.Header.Get(mcpSessionHeader)
	if sessionID != "" {
		s.store.MarkInitialized(sessionID)
	}
	w.WriteHeader(http.StatusAccepted)
}
```

**Step 6: Update `dispatchToolCall` to use session-aware vault resolution**

Replace the `vaultFromArgs` call and error at the top of `dispatchToolCall` with:

```go
func (s *MCPServer) dispatchToolCall(ctx context.Context, w http.ResponseWriter, req *JSONRPCRequest, sess *mcpSession) {
	if req.Params == nil {
		sendError(w, req.ID, -32602, "invalid params: params required")
		return
	}
	args := req.Params.Arguments
	if args == nil {
		args = make(map[string]any)
	}

	// Check initialized gate
	if sess != nil && !sess.initialized.Load() {
		sendError(w, req.ID, -32002, "session not initialized: send 'notifications/initialized' first")
		return
	}

	vault, errMsg := resolveVault(sess, args)
	if errMsg != "" {
		sendError(w, req.ID, -32603, errMsg) // -32603 = vault mismatch
		return
	}

	// ... rest of handler dispatch unchanged ...
}
```

Update `handleRPC` to pass session into `dispatchToolCall`:
```go
case "tools/call":
    sess, _ := sessionFromRequest(r, s.store)
    s.dispatchToolCall(ctx, w, &req, sess)
```

**Step 7: Add fan-out goroutine**

```go
// fanOutLoop reads from the EventBus and dispatches to SSE-connected sessions.
// Runs until the EventBus channel is closed (in Shutdown).
func (s *MCPServer) fanOutLoop() {
	defer close(s.fanDone)
	for event := range s.bus.Events() {
		sessions := s.store.ByVault(event.Vault)
		for _, sess := range sessions {
			sess := sess // capture for goroutine
			go func() {
				if !sess.streaming.Load() {
					return
				}
				select {
				case sess.pushCh <- event:
				default:
					sess.droppedEvents.Add(1)
				}
			}()
		}
	}
}
```

**Step 8: Write tests for initialize flow**

Add to `internal/mcp/server_test.go`:

```go
func TestHandleInitialize(t *testing.T) {
	s := newTestServer(t)
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
	sessionID := w.Header().Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatal("expected Mcp-Session-Id header")
	}
	if len(sessionID) != 64 { // 32 bytes hex = 64 chars
		t.Fatalf("session ID wrong length: %d", len(sessionID))
	}
}

func TestHandleInitialize_VaultPinned(t *testing.T) {
	s := newTestServer(t)
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp?vault=my-project", strings.NewReader(body))
	w := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(w, req)

	sessionID := w.Header().Get("Mcp-Session-Id")
	sess, ok := s.store.Get(sessionID)
	if !ok {
		t.Fatal("session not found")
	}
	if sess.vault != "my-project" {
		t.Fatalf("vault: got %q, want %q", sess.vault, "my-project")
	}
}

func TestToolCallRejectsBeforeInitialized(t *testing.T) {
	s := newTestServer(t)
	// create session but don't send initialized notification
	h := [32]byte{}
	id, _ := s.store.Create("v", h)

	body := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"muninn_status","arguments":{}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Mcp-Session-Id", id)
	w := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(w, req)

	var resp JSONRPCResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32002 {
		t.Fatalf("expected -32002, got %+v", resp.Error)
	}
}
```

Note: `newTestServer` needs to be updated to expose `store` field — see current `server_test.go` for the helper pattern.

**Step 9: Run tests**

```bash
go test ./internal/mcp/... -v -race
```
Expected: all pass.

**Step 10: Commit**

```bash
git add internal/mcp/server.go internal/mcp/server_test.go
git commit -m "feat(mcp): initialize/initialized handshake and session-aware dispatch"
```

---

## Task 7: MCP server — SSE GET endpoint

**Files:**
- Modify: `internal/mcp/server.go` (add `handleSSE`)

**Step 1: Implement `handleSSE`**

Add this method to server.go:

```go
// handleSSE serves the SSE stream for server-to-client cognitive push notifications.
// Requires: Accept: text/event-stream, Authorization: Bearer <token>, Mcp-Session-Id: <id>
func (s *MCPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Validate Accept header per MCP spec
	if !strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		http.Error(w, "Accept: text/event-stream required", http.StatusMethodNotAllowed)
		return
	}

	// Auth
	auth := authFromRequest(r, s.token)
	if !auth.Authorized {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Session lookup
	sess, sessionID := sessionFromRequest(r, s.store)
	if sessionID == "" {
		http.Error(w, "Mcp-Session-Id header required", http.StatusBadRequest)
		return
	}
	if sess == nil {
		http.Error(w, "session not found or expired — re-initialize", http.StatusBadRequest)
		return
	}

	// Token binding verification
	if s.token != "" {
		if msg := validateSessionToken(sess, auth.Token); msg != "" {
			http.Error(w, msg, http.StatusUnauthorized)
			return
		}
	}

	// Initialized gate
	if !sess.initialized.Load() {
		http.Error(w, "session not initialized", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	s.store.Touch(sessionID)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sess.streaming.Store(true)
	defer sess.streaming.Store(false)

	for {
		select {
		case event, ok := <-sess.pushCh:
			if !ok {
				// pushCh closed by reaper (session expired)
				return
			}
			data, err := json.Marshal(map[string]any{
				"jsonrpc": "2.0",
				"method":  event.Method,
				"params":  event.Params,
			})
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
```

**Step 2: Write SSE test**

```go
func TestHandleSSE_RequiresAcceptHeader(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	// no Accept: text/event-stream
	w := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got %d, want 405", w.Code)
	}
}

func TestHandleSSE_RequiresInitializedSession(t *testing.T) {
	s := newTestServer(t)
	h := [32]byte{}
	id, _ := s.store.Create("v", h)
	// session exists but not initialized

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Mcp-Session-Id", id)
	w := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
}
```

**Step 3: Run tests**

```bash
go test ./internal/mcp/... -v -race
```
Expected: all pass.

**Step 4: Commit**

```bash
git add internal/mcp/server.go internal/mcp/server_test.go
git commit -m "feat(mcp): SSE GET /mcp endpoint for cognitive push notifications"
```

---

## Task 8: Engine — inject EventBus

**Files:**
- Modify: `internal/engine/engine.go`

Add `EventBus` as an optional field on `Engine`. The engine passes it to the trigger system and cognitive workers so they can emit events.

**Step 1: Add the field**

In the `Engine` struct, after the `scoring` field, add:

```go
// eventBus routes cognitive events to active MCP sessions. nil-safe (use NoopEventBus{} for tests).
eventBus mcp.EventBus
```

Add import: `"github.com/scrypster/muninndb/internal/mcp"` to engine.go imports.

**Step 2: Find the `New` or constructor function for Engine and add EventBus parameter**

Search for the Engine constructor:

```bash
grep -n "func New\b" internal/engine/engine.go
```

Add `bus mcp.EventBus` as the last parameter. If nil is passed, default to `mcp.NoopEventBus{}`:

```go
if bus == nil {
    bus = mcp.NoopEventBus{}
}
eng.eventBus = bus
```

**Step 3: Pass eventBus to trigger system**

Find where `trigger.NewTriggerSystem(...)` is called in engine.go and add the bus. You'll need to update `TriggerSystem` in the next task to accept it — for now, just store it on the engine.

**Step 4: Update `cmd/muninn/server.go`** to pass the EventBus to the Engine constructor and to the MCP server:

```go
bus := mcp.NewEventBus(256)
eng := engine.New(..., bus)
mcpSrv := mcp.New(mcpAddr, eng, token, bus)
```

**Step 5: Build verify**

```bash
go build ./...
```
Expected: no errors.

**Step 6: Commit**

```bash
git add internal/engine/engine.go cmd/muninn/server.go
git commit -m "feat(engine): inject EventBus for cognitive push routing"
```

---

## Task 9: Trigger system — emit contradiction + activation events

**Files:**
- Modify: `internal/engine/trigger/system.go`

The trigger system already has `TriggerType` constants `TriggerContradiction` and `TriggerThresholdCrossed`. Add EventBus field and emit events at the appropriate dispatch points.

**Step 1: Read the trigger system's dispatch/deliver section**

```bash
grep -n "TriggerContradiction\|TriggerThresholdCrossed\|DeliverFunc\|dispatch\|deliver" internal/engine/trigger/system.go | head -40
```

**Step 2: Add EventBus to `TriggerSystem` struct**

```go
type TriggerSystem struct {
    // ... existing fields ...
    eventBus EventBusEmitter // optional; nil = no push
}

// EventBusEmitter is the subset of mcp.EventBus needed by the trigger system.
// Using a local interface avoids an import cycle.
type EventBusEmitter interface {
    Publish(vault, method string, params []byte)
}
```

Wait — to avoid import cycles (trigger importing mcp, mcp importing trigger would be a cycle since engine imports both), define a minimal local interface in the trigger package:

```go
// EventEmitter is implemented by mcp.EventBus. Defined here to avoid import cycles.
type EventEmitter interface {
    PublishRaw(vault, method string, params []byte)
}
```

Then update `mcp.EventBus` and `chanEventBus` to also implement `PublishRaw`. Or simpler: use a callback function:

```go
// PushFunc is called when the trigger system wants to push a cognitive event.
// vault: the vault where the event occurred.
// method: MCP notification method name.
// params: pre-serialized JSON params.
type PushFunc func(vault, method string, params []byte)
```

Add `pushFn PushFunc` to `TriggerSystem`. In engine.go, wire it up:

```go
pushFn := func(vault, method string, params []byte) {
    eng.eventBus.Publish(mcp.MCPEvent{
        Vault:  vault,
        Method: method,
        Params: json.RawMessage(params),
    })
}
triggers := trigger.New(..., pushFn)
```

**Step 3: Find where contradiction and threshold events are dispatched in trigger system**

```bash
grep -n "TriggerContradiction\|TriggerThresholdCrossed\|Deliver\|push" internal/engine/trigger/system.go | head -40
```

At those dispatch points, add:

```go
// Emit to MCP push (non-blocking — engine must never block on AI client)
if s.pushFn != nil {
    params, _ := json.Marshal(mcp.ContradictionParams{IDa: push.Engram.ID, IDb: otherID})
    s.pushFn(vault, "notifications/muninn/contradiction", params)
}
```

For threshold-crossed:
```go
if s.pushFn != nil {
    params, _ := json.Marshal(mcp.ActivationParams{
        ID:      push.Engram.ID,
        Concept: push.Engram.Concept,
        Score:   push.Score,
        Vault:   vault,
    })
    s.pushFn(vault, "notifications/muninn/activation", params)
}
```

**Step 4: Build verify**

```bash
go build ./...
```

**Step 5: Commit**

```bash
git add internal/engine/trigger/system.go internal/engine/engine.go
git commit -m "feat(trigger): emit contradiction and activation events to EventBus"
```

---

## Task 10: Cognitive worker — emit Hebbian association events

**Files:**
- Modify: `internal/cognitive/worker.go` or the HebbianWorker implementation

**Step 1: Find where Hebbian associations are persisted**

```bash
grep -rn "Hebbian\|hebbianWorker\|assoc\|Weight\|Link" internal/cognitive/ | grep -v "_test" | head -30
```

**Step 2: Add `PushFunc` to HebbianWorker**

Same pattern as trigger system — add a `pushFn` callback field. Wire from engine.go.

**Step 3: Emit association event when a new link is strengthened past a threshold (e.g., weight > 0.7)**

```go
if eng.pushFn != nil && newWeight > 0.7 {
    params, _ := json.Marshal(mcp.AssociationParams{
        SourceID: sourceID,
        TargetID: targetID,
        Weight:   newWeight,
    })
    eng.pushFn(vault, "notifications/muninn/association", params)
}
```

**Step 4: Build verify + run tests**

```bash
go build ./... && go test ./internal/cognitive/... -race
```

**Step 5: Commit**

```bash
git add internal/cognitive/
git commit -m "feat(cognitive): emit Hebbian association events to EventBus"
```

---

## Task 11: Admin endpoint — GET /api/admin/mcp/sessions

**Files:**
- Modify: `internal/transport/rest/admin_handlers.go`
- Modify: `internal/transport/rest/server.go` (add route + pass session store ref)

The REST server needs a reference to the session store to expose session metrics. Add it as a field on the REST `Server` struct.

**Step 1: Add `mcpSessionStore` to REST Server struct**

In `internal/transport/rest/server.go`:

```go
type Server struct {
    // ... existing fields ...
    mcpSessionStore interface {
        // Minimal interface for admin visibility — avoids importing internal/mcp
        Sessions() []MCPSessionInfo
        BusDropped() int64
    }
}
```

Or simpler — define a `MCPSessionLister` interface in the REST package and have the MCP session store implement it. Actually the cleanest approach: define a `MCPStats` struct and a `MCPStatser` interface in REST:

```go
// MCPSessionInfo is the admin view of a single MCP session.
type MCPSessionInfo struct {
    SessionID     string    `json:"session_id"`
    Vault         string    `json:"vault"`
    Initialized   bool      `json:"initialized"`
    Streaming     bool      `json:"streaming"`
    DroppedEvents int64     `json:"dropped_events"`
    CreatedAt     time.Time `json:"created_at"`
    LastUsed      time.Time `json:"last_used"`
}
```

Pass a callback or interface from `cmd/muninn/server.go` that returns `[]MCPSessionInfo`.

**Step 2: Add the handler**

```go
func (s *Server) handleMCPSessions(w http.ResponseWriter, r *http.Request) {
    if s.mcpStatsFn == nil {
        s.sendJSON(w, http.StatusOK, map[string]any{
            "sessions":          []any{},
            "session_count":     0,
            "session_cap":       256,
            "bus_dropped_events": 0,
        })
        return
    }
    stats := s.mcpStatsFn()
    s.sendJSON(w, http.StatusOK, stats)
}
```

**Step 3: Add route in `NewServer`**

```go
mux.Handle("GET /api/admin/mcp/sessions", adminAuth(http.HandlerFunc(s.handleMCPSessions)))
```

**Step 4: Build verify + run tests**

```bash
go build ./... && go test ./internal/transport/rest/... -race
```

**Step 5: Commit**

```bash
git add internal/transport/rest/admin_handlers.go internal/transport/rest/server.go cmd/muninn/server.go
git commit -m "feat(admin): GET /api/admin/mcp/sessions for session observability"
```

---

## Task 12: Final integration test + build

**Step 1: Run all tests**

```bash
go test ./... -race -count=1
```
Expected: all pass.

**Step 2: Build all targets**

```bash
go build ./...
```
Expected: no errors.

**Step 3: Manual smoke test** (requires running server)

```bash
# Start server
./muninn start

# Test initialize
curl -s -X POST http://localhost:8750/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  -v 2>&1 | grep -E "Mcp-Session-Id|result"

# Test with vault-pinned URL
curl -s -X POST "http://localhost:8750/mcp?vault=my-project" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"cursor","version":"1.0"}}}' \
  -v 2>&1 | grep "Mcp-Session-Id"

# Test admin sessions endpoint
curl -s http://localhost:8476/api/admin/mcp/sessions | jq .
```

**Step 4: Final commit**

```bash
git commit --allow-empty -m "feat: MCP sessions, vault pinning, SSE push — complete"
```

---

## Reference: Cursor project config (what users configure)

After this is live, Cursor project `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "muninn": {
      "url": "http://localhost:8750/mcp?vault=my-project",
      "headers": {
        "Authorization": "Bearer <contents of ~/.muninn/mcp.token>"
      }
    }
  }
}
```

Global Claude Desktop / Claude Code (no vault pin, uses `"default"` or explicit arg):

```json
{
  "mcpServers": {
    "muninn": {
      "url": "http://localhost:8750/mcp"
    }
  }
}
```
