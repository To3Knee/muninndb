# Plasticity Configuration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Add per-vault cognitive pipeline configuration ("Plasticity") — presets and optional overrides for decay, Hebbian learning, BFS hops, and scoring weights — persisted in `VaultConfig` and exposed in the vault settings web UI.

**Architecture:** A `PlasticityConfig` pointer field is added to `VaultConfig` (nil = use preset defaults). `ResolvePlasticity()` merges any overrides atop the chosen preset. The Engine resolves it in `Activate()` and gates Hebbian/Decay worker submissions accordingly. The REST layer exposes `GET/PUT /api/admin/vault/{name}/plasticity`. The web UI adds an Alpine.js card in the vault settings tab.

**Tech Stack:** Go (internal/auth, internal/engine, internal/transport/rest), Alpine.js (web/static/js/app.js), Pebble KV (VaultConfig persistence)

---

## Naming Note

The term **"engram"** in this codebase refers to a single memory unit (see `docs/engram.md`). This feature is called **"Plasticity"** — the per-vault cognitive pipeline configuration. All types, endpoints, and UI labels use "Plasticity" / "plasticity".

---

## Task 1 — Data model: PlasticityConfig + ResolvedPlasticity in internal/auth

**Files:**
- Modify: `internal/auth/types.go`
- Create: `internal/auth/plasticity.go`

**Step 1a: Add PlasticityConfig to VaultConfig in internal/auth/types.go**

Find `VaultConfig` struct and add the `Plasticity` pointer field:

```go
type VaultConfig struct {
    Name       string           `json:"name"`
    Public     bool             `json:"public"`
    CreatedAt  time.Time        `json:"created_at"`
    Plasticity *PlasticityConfig `json:"plasticity,omitempty"` // NEW
}
```

**Step 1b: Create internal/auth/plasticity.go**

```go
package auth

// PlasticityConfig is the per-vault cognitive pipeline configuration.
// A nil PlasticityConfig means "use defaults" (equivalent to Preset: "default").
// Non-nil pointer fields override the chosen preset value.
type PlasticityConfig struct {
    Version int    `json:"version,omitempty"` // schema version, currently 1
    Preset  string `json:"preset,omitempty"`  // "default" | "reference" | "scratchpad" | "knowledge-graph"

    // Optional overrides (nil = use preset value)
    HebbianEnabled *bool    `json:"hebbian_enabled,omitempty"`
    DecayEnabled   *bool    `json:"decay_enabled,omitempty"`
    HopDepth       *int     `json:"hop_depth,omitempty"`       // BFS hops 0–8
    SemanticWeight *float32 `json:"semantic_weight,omitempty"` // 0–1
    FTSWeight      *float32 `json:"fts_weight,omitempty"`      // 0–1
    DecayFloor     *float32 `json:"decay_floor,omitempty"`     // 0–1
    DecayStability *float32 `json:"decay_stability,omitempty"` // days
}

// ResolvedPlasticity is the fully-merged configuration after applying preset defaults
// and any field-level overrides from PlasticityConfig.
type ResolvedPlasticity struct {
    HebbianEnabled bool
    DecayEnabled   bool
    HopDepth       int
    SemanticWeight float32
    FTSWeight      float32
    DecayFloor     float32
    DecayStability float32 // days
    HebbianWeight  float32
    DecayWeight    float32
    RecencyWeight  float32
}

type plasticityPreset struct {
    HebbianEnabled bool
    DecayEnabled   bool
    HopDepth       int
    SemanticWeight float32
    FTSWeight      float32
    DecayFloor     float32
    DecayStability float32
    HebbianWeight  float32
    DecayWeight    float32
    RecencyWeight  float32
}

var plasticityPresets = map[string]plasticityPreset{
    "default": {
        HebbianEnabled: true,
        DecayEnabled:   true,
        HopDepth:       2,
        SemanticWeight: 0.6,
        FTSWeight:      0.3,
        DecayFloor:     0.05,
        DecayStability: 30,
        HebbianWeight:  0.5,
        DecayWeight:    0.4,
        RecencyWeight:  0.3,
    },
    "reference": {
        HebbianEnabled: true,
        DecayEnabled:   false, // memories persist indefinitely
        HopDepth:       3,
        SemanticWeight: 0.7,
        FTSWeight:      0.5,
        DecayFloor:     1.0, // relevance never falls
        DecayStability: 365,
        HebbianWeight:  0.6,
        DecayWeight:    0.0,
        RecencyWeight:  0.1,
    },
    "scratchpad": {
        HebbianEnabled: false,
        DecayEnabled:   true,
        HopDepth:       0,
        SemanticWeight: 0.5,
        FTSWeight:      0.4,
        DecayFloor:     0.01,
        DecayStability: 7,
        HebbianWeight:  0.0,
        DecayWeight:    0.8,
        RecencyWeight:  0.5,
    },
    "knowledge-graph": {
        HebbianEnabled: true,
        DecayEnabled:   true,
        HopDepth:       4,
        SemanticWeight: 0.5,
        FTSWeight:      0.2,
        DecayFloor:     0.1,
        DecayStability: 60,
        HebbianWeight:  0.8,
        DecayWeight:    0.2,
        RecencyWeight:  0.2,
    },
}

// ResolvePlasticity merges cfg (which may be nil) atop its chosen preset,
// returning a fully-populated ResolvedPlasticity.
func ResolvePlasticity(cfg *PlasticityConfig) ResolvedPlasticity {
    presetName := "default"
    if cfg != nil && cfg.Preset != "" {
        presetName = cfg.Preset
    }
    p, ok := plasticityPresets[presetName]
    if !ok {
        p = plasticityPresets["default"]
    }

    r := ResolvedPlasticity{
        HebbianEnabled: p.HebbianEnabled,
        DecayEnabled:   p.DecayEnabled,
        HopDepth:       p.HopDepth,
        SemanticWeight: p.SemanticWeight,
        FTSWeight:      p.FTSWeight,
        DecayFloor:     p.DecayFloor,
        DecayStability: p.DecayStability,
        HebbianWeight:  p.HebbianWeight,
        DecayWeight:    p.DecayWeight,
        RecencyWeight:  p.RecencyWeight,
    }

    if cfg == nil {
        return r
    }

    // Apply pointer-field overrides
    if cfg.HebbianEnabled != nil {
        r.HebbianEnabled = *cfg.HebbianEnabled
    }
    if cfg.DecayEnabled != nil {
        r.DecayEnabled = *cfg.DecayEnabled
    }
    if cfg.HopDepth != nil {
        r.HopDepth = *cfg.HopDepth
    }
    if cfg.SemanticWeight != nil {
        r.SemanticWeight = *cfg.SemanticWeight
    }
    if cfg.FTSWeight != nil {
        r.FTSWeight = *cfg.FTSWeight
    }
    if cfg.DecayFloor != nil {
        r.DecayFloor = *cfg.DecayFloor
    }
    if cfg.DecayStability != nil {
        r.DecayStability = *cfg.DecayStability
    }

    return r
}

// ValidPlasticityPreset returns true if s is a known preset name.
func ValidPlasticityPreset(s string) bool {
    _, ok := plasticityPresets[s]
    return ok
}
```

**Step 1c: Write failing test**

Create `internal/auth/plasticity_test.go`:

```go
package auth

import (
    "testing"
)

func TestResolvePlasticity_NilUsesDefault(t *testing.T) {
    r := ResolvePlasticity(nil)
    if r.HopDepth != 2 {
        t.Errorf("want HopDepth=2, got %d", r.HopDepth)
    }
    if !r.HebbianEnabled {
        t.Error("want HebbianEnabled=true")
    }
    if !r.DecayEnabled {
        t.Error("want DecayEnabled=true")
    }
}

func TestResolvePlasticity_ScratchpadPreset(t *testing.T) {
    r := ResolvePlasticity(&PlasticityConfig{Preset: "scratchpad"})
    if r.HopDepth != 0 {
        t.Errorf("scratchpad HopDepth want 0, got %d", r.HopDepth)
    }
    if r.HebbianEnabled {
        t.Error("scratchpad: want HebbianEnabled=false")
    }
    if r.DecayStability != 7 {
        t.Errorf("scratchpad DecayStability want 7, got %f", r.DecayStability)
    }
}

func TestResolvePlasticity_ReferencePreset(t *testing.T) {
    r := ResolvePlasticity(&PlasticityConfig{Preset: "reference"})
    if r.DecayEnabled {
        t.Error("reference: want DecayEnabled=false")
    }
    if r.DecayFloor != 1.0 {
        t.Errorf("reference DecayFloor want 1.0, got %f", r.DecayFloor)
    }
}

func TestResolvePlasticity_KnowledgeGraphPreset(t *testing.T) {
    r := ResolvePlasticity(&PlasticityConfig{Preset: "knowledge-graph"})
    if r.HopDepth != 4 {
        t.Errorf("knowledge-graph HopDepth want 4, got %d", r.HopDepth)
    }
}

func TestResolvePlasticity_PointerOverride(t *testing.T) {
    hd := 5
    r := ResolvePlasticity(&PlasticityConfig{
        Preset:   "default",
        HopDepth: &hd,
    })
    if r.HopDepth != 5 {
        t.Errorf("override HopDepth want 5, got %d", r.HopDepth)
    }
    // Other default fields still apply
    if !r.HebbianEnabled {
        t.Error("want HebbianEnabled=true (from default)")
    }
}

func TestResolvePlasticity_BoolOverride(t *testing.T) {
    f := false
    r := ResolvePlasticity(&PlasticityConfig{
        Preset:         "default",
        HebbianEnabled: &f,
    })
    if r.HebbianEnabled {
        t.Error("explicit false override should set HebbianEnabled=false")
    }
}

func TestResolvePlasticity_InvalidPresetFallsToDefault(t *testing.T) {
    r := ResolvePlasticity(&PlasticityConfig{Preset: "bogus"})
    if r.HopDepth != 2 {
        t.Errorf("invalid preset should fall to default, want HopDepth=2, got %d", r.HopDepth)
    }
}

func TestValidPlasticityPreset(t *testing.T) {
    if !ValidPlasticityPreset("default") { t.Error("default should be valid") }
    if !ValidPlasticityPreset("reference") { t.Error("reference should be valid") }
    if !ValidPlasticityPreset("scratchpad") { t.Error("scratchpad should be valid") }
    if !ValidPlasticityPreset("knowledge-graph") { t.Error("knowledge-graph should be valid") }
    if ValidPlasticityPreset("bogus") { t.Error("bogus should not be valid") }
}
```

**Step 1d: Run failing tests**

```bash
cd /Users/mjbonanno/github.com/scrypster/muninndb
go test ./internal/auth/... -run TestResolvePlasticity -v
```

Expected: FAIL — `PlasticityConfig` undefined.

**Step 1e: Implement (create the files above)**

**Step 1f: Run tests again**

```bash
go test ./internal/auth/... -v
```

Expected: PASS.

**Step 1g: Commit**

```bash
git add internal/auth/plasticity.go internal/auth/plasticity_test.go internal/auth/types.go
git commit -m "feat(plasticity): PlasticityConfig + ResolvedPlasticity data model with presets and overrides"
```

---

## Task 2 — REST: GET/PUT /api/admin/vault/{name}/plasticity

**Files:**
- Modify: `internal/transport/rest/admin_handlers.go`
- Modify: `internal/transport/rest/server.go`
- Create: `internal/transport/rest/admin_plasticity_test.go`

**Step 2a: Write failing tests first**

Create `internal/transport/rest/admin_plasticity_test.go`:

```go
package rest

import (
    "bytes"
    "encoding/json"
    "net/http/httptest"
    "testing"

    "github.com/scrypster/muninndb/internal/auth"
)

func TestHandleGetVaultPlasticity_DefaultWhenNil(t *testing.T) {
    as := newTestAuthStore(t)
    server := NewServer("localhost:0", &MockEngine{}, as, nil, nil, EmbedInfo{}, nil)

    req := httptest.NewRequest("GET", "/api/admin/vault/myvault/plasticity", nil)
    req.SetPathValue("name", "myvault")
    w := httptest.NewRecorder()
    server.handleGetVaultPlasticity(as)(w, req)

    if w.Code != 200 {
        t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
    }
    var resp struct {
        Raw      *auth.PlasticityConfig  `json:"config"`
        Resolved auth.ResolvedPlasticity `json:"resolved"`
    }
    if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
        t.Fatalf("decode: %v", err)
    }
    if resp.Raw != nil {
        t.Error("config should be nil when no Plasticity set")
    }
    if resp.Resolved.HopDepth != 2 {
        t.Errorf("resolved HopDepth want 2, got %d", resp.Resolved.HopDepth)
    }
}

func TestHandleGetVaultPlasticity_MissingName(t *testing.T) {
    as := newTestAuthStore(t)
    server := NewServer("localhost:0", &MockEngine{}, as, nil, nil, EmbedInfo{}, nil)

    req := httptest.NewRequest("GET", "/api/admin/vault//plasticity", nil)
    req.SetPathValue("name", "")
    w := httptest.NewRecorder()
    server.handleGetVaultPlasticity(as)(w, req)

    if w.Code != 400 {
        t.Errorf("expected 400, got %d", w.Code)
    }
}

func TestHandlePutVaultPlasticity_RoundTrip(t *testing.T) {
    as := newTestAuthStore(t)
    server := NewServer("localhost:0", &MockEngine{}, as, nil, nil, EmbedInfo{}, nil)

    body, _ := json.Marshal(auth.PlasticityConfig{Preset: "scratchpad"})
    req := httptest.NewRequest("PUT", "/api/admin/vault/myvault/plasticity", bytes.NewReader(body))
    req.SetPathValue("name", "myvault")
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    server.handlePutVaultPlasticity(as)(w, req)

    if w.Code != 200 {
        t.Fatalf("PUT: expected 200, got %d: %s", w.Code, w.Body.String())
    }

    // GET and verify persisted
    req2 := httptest.NewRequest("GET", "/api/admin/vault/myvault/plasticity", nil)
    req2.SetPathValue("name", "myvault")
    w2 := httptest.NewRecorder()
    server.handleGetVaultPlasticity(as)(w2, req2)

    var resp struct {
        Raw      *auth.PlasticityConfig  `json:"config"`
        Resolved auth.ResolvedPlasticity `json:"resolved"`
    }
    json.NewDecoder(w2.Body).Decode(&resp)
    if resp.Raw == nil || resp.Raw.Preset != "scratchpad" {
        t.Errorf("expected scratchpad preset, got %+v", resp.Raw)
    }
    if resp.Resolved.HopDepth != 0 {
        t.Errorf("scratchpad HopDepth want 0, got %d", resp.Resolved.HopDepth)
    }
}

func TestHandlePutVaultPlasticity_InvalidPreset(t *testing.T) {
    as := newTestAuthStore(t)
    server := NewServer("localhost:0", &MockEngine{}, as, nil, nil, EmbedInfo{}, nil)

    body, _ := json.Marshal(auth.PlasticityConfig{Preset: "invalid"})
    req := httptest.NewRequest("PUT", "/api/admin/vault/myvault/plasticity", bytes.NewReader(body))
    req.SetPathValue("name", "myvault")
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    server.handlePutVaultPlasticity(as)(w, req)

    if w.Code != 400 {
        t.Errorf("expected 400, got %d", w.Code)
    }
}
```

**Step 2b: Run failing tests**

```bash
go test ./internal/transport/rest/... -run TestHandleVaultPlasticity -v
```

Expected: FAIL — `handleGetVaultPlasticity` undefined.

**Step 2c: Add handlers to admin_handlers.go**

Find the pattern of existing handlers like `handleSetVaultConfig`. Add after the last admin handler:

```go
// handleGetVaultPlasticity returns the raw PlasticityConfig (may be nil) and
// the fully-resolved config for the named vault.
func (s *Server) handleGetVaultPlasticity(as *auth.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        name := r.PathValue("name")
        if name == "" {
            http.Error(w, "vault name required", http.StatusBadRequest)
            return
        }
        vc, err := as.GetVaultConfig(name)
        if err != nil {
            // Vault not found — return defaults
            vc = auth.VaultConfig{Name: name}
        }
        resolved := auth.ResolvePlasticity(vc.Plasticity)
        writeJSON(w, map[string]any{
            "config":   vc.Plasticity,
            "resolved": resolved,
        })
    }
}

// handlePutVaultPlasticity updates the PlasticityConfig for the named vault.
func (s *Server) handlePutVaultPlasticity(as *auth.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        name := r.PathValue("name")
        if name == "" {
            http.Error(w, "vault name required", http.StatusBadRequest)
            return
        }
        var cfg auth.PlasticityConfig
        if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
            http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
            return
        }
        preset := cfg.Preset
        if preset == "" {
            preset = "default"
        }
        if !auth.ValidPlasticityPreset(preset) {
            http.Error(w, "unknown preset: "+preset, http.StatusBadRequest)
            return
        }
        cfg.Preset = preset

        vc, err := as.GetVaultConfig(name)
        if err != nil {
            vc = auth.VaultConfig{Name: name}
        }
        vc.Plasticity = &cfg
        if err := as.SetVaultConfig(vc); err != nil {
            http.Error(w, "store error: "+err.Error(), http.StatusInternalServerError)
            return
        }
        resolved := auth.ResolvePlasticity(&cfg)
        writeJSON(w, map[string]any{
            "config":   &cfg,
            "resolved": resolved,
        })
    }
}
```

**Step 2d: Register routes in server.go**

Find the `mux.HandleFunc` block for admin routes and add:

```go
mux.HandleFunc("GET /api/admin/vault/{name}/plasticity", adminOnly(s.handleGetVaultPlasticity(authStore)))
mux.HandleFunc("PUT /api/admin/vault/{name}/plasticity", adminOnly(s.handlePutVaultPlasticity(authStore)))
```

**Step 2e: Run tests**

```bash
go test ./internal/transport/rest/... -v
```

Expected: PASS.

**Step 2f: Commit**

```bash
git add internal/transport/rest/admin_handlers.go internal/transport/rest/server.go internal/transport/rest/admin_plasticity_test.go
git commit -m "feat(plasticity): REST GET/PUT /api/admin/vault/{name}/plasticity handlers + tests"
```

---

## Task 3 — Engine integration: resolve Plasticity in Activate(), gate workers

**Files:**
- Modify: `internal/engine/engine.go`
- Modify: `internal/engine/engine_test.go`

**Context:** `NewEngine` call sites are:
- `internal/engine/engine_test.go` (2 locations)
- `cmd/eval/main.go`
- `cmd/muninn/server.go`
- `cmd/bench/main.go`

**Step 3a: Add authStore field to Engine struct**

In `internal/engine/engine.go`, in the `Engine` struct definition, add:

```go
authStore *auth.Store // for per-vault Plasticity config lookup; nil = use defaults
```

**Step 3b: Add authStore parameter to NewEngine**

```go
func NewEngine(
    store *storage.PebbleStore,
    authStore *auth.Store,     // NEW — insert after store
    ftsIdx *fts.Index,
    // ... rest unchanged
) *Engine {
    e := &Engine{
        store:     store,
        authStore: authStore,  // NEW
        // ... rest unchanged
```

**Step 3c: Resolve Plasticity at the top of Activate()**

At the top of the `Activate()` method, before building `actReq`, add:

```go
// Resolve per-vault Plasticity config (nil authStore = use defaults, e.g. in tests)
var resolved auth.ResolvedPlasticity
if e.authStore != nil {
    vaultCfg, err := e.authStore.GetVaultConfig(req.Vault)
    if err == nil {
        resolved = auth.ResolvePlasticity(vaultCfg.Plasticity)
    } else {
        resolved = auth.ResolvePlasticity(nil)
    }
} else {
    resolved = auth.ResolvePlasticity(nil)
}
```

**Step 3d: Use resolved.HopDepth instead of hard-coded 2**

Find the block:
```go
if actReq.HopDepth == 0 {
    actReq.HopDepth = 2
}
```

Replace with:
```go
if actReq.HopDepth == 0 {
    actReq.HopDepth = resolved.HopDepth
}
```

**Step 3e: Apply preset weights when client provides none**

Find the existing `if req.Weights != nil { ... }` block and add an else clause:

```go
} else {
    actReq.Weights = &activation.Weights{
        SemanticSimilarity: float64(resolved.SemanticWeight),
        FullTextRelevance:  float64(resolved.FTSWeight),
        DecayFactor:        float64(resolved.DecayWeight),
        HebbianBoost:       float64(resolved.HebbianWeight),
        Recency:            float64(resolved.RecencyWeight),
    }
}
```

**Step 3f: Gate Hebbian worker submission**

Find:
```go
if e.hebbianWorker != nil && len(result.Activations) > 0 && !auth.ObserveFromContext(ctx) {
```

Replace with:
```go
if e.hebbianWorker != nil && len(result.Activations) > 0 && !auth.ObserveFromContext(ctx) && resolved.HebbianEnabled {
```

**Step 3g: Gate Decay worker submission**

Find:
```go
if e.decayWorker != nil && len(result.Activations) > 0 && !auth.ObserveFromContext(ctx) {
```

Replace with:
```go
if e.decayWorker != nil && len(result.Activations) > 0 && !auth.ObserveFromContext(ctx) && resolved.DecayEnabled {
```

**Step 3h: Update call sites**

In `internal/engine/engine_test.go` (2 locations): insert `nil` as second argument after `store`:
```go
eng := NewEngine(store, nil, ftsIdx, actEngine, trigSystem, ...)
```

In `cmd/muninn/server.go`: pass the auth store (it's already in scope as `authStore`):
```go
eng := engine.NewEngine(store, authStore, ftsIndex, actEngine, trigSystem, ...)
```

In `cmd/eval/main.go`: pass the auth store (already in scope):
```go
eng := engine.NewEngine(store, authStore, ftsIdx, actEngine, trigSystem, ...)
```

In `cmd/bench/main.go`: pass `nil` (bench has no auth store):
```go
eng := engine.NewEngine(store, nil, ftsIdx, actEngine, trigSystem, ...)
```

**Step 3i: Add a test for Plasticity gating**

In `internal/engine/engine_test.go`, add:

```go
func TestActivate_PlasticityGatesHebbian(t *testing.T) {
    // scratchpad preset: HebbianEnabled=false, HopDepth=0
    dir := t.TempDir()
    as, err := auth.NewStore(dir)
    if err != nil {
        t.Fatalf("auth.NewStore: %v", err)
    }
    t.Cleanup(func() { as.Close() })
    _ = as.SetVaultConfig(auth.VaultConfig{
        Name:   "testvault",
        Public: true,
        Plasticity: &auth.PlasticityConfig{Preset: "scratchpad"},
    })

    eng, cleanup := newTestEnv(t)
    defer cleanup()
    // Inject the auth store into the engine
    eng.authStore = as

    ctx := context.Background()
    resp, err := eng.Activate(ctx, &mbp.ActivateRequest{Vault: "testvault", Context: []string{"test"}})
    if err != nil {
        t.Fatalf("Activate: %v", err)
    }
    _ = resp // No panic, HopDepth=0, HebbianEnabled=false
}
```

**Step 3j: Run all engine tests**

```bash
go test ./internal/engine/... -v -timeout 60s
```

Expected: PASS.

**Step 3k: Build check**

```bash
go build ./...
```

Expected: no errors.

**Step 3l: Commit**

```bash
git add internal/engine/engine.go internal/engine/engine_test.go cmd/muninn/server.go cmd/eval/main.go cmd/bench/main.go
git commit -m "feat(plasticity): engine resolves per-vault PlasticityConfig in Activate(), gates Hebbian/Decay workers"
```

---

## Task 4 — Web UI: Plasticity card in vault settings tab

**Files:**
- Modify: `web/templates/index.html`
- Modify: `web/static/js/app.js`

**Step 4a: Add Alpine.js state to app.js**

Find the block with `cogWorkerStats: null,` and add below it:

```js
// Plasticity (vault cognitive pipeline config)
plasticityForm: {
  preset: 'default',
  showAdvanced: false,
  hebbianEnabled: true,
  decayEnabled: true,
  hopDepth: null,
  semanticWeight: null,
  ftsWeight: null,
  decayFloor: null,
  decayStability: null,
},
plasticitySaving: false,
plasticitySaveOk: false,
plasticitySaveErr: '',
```

**Step 4b: Trigger loadPlasticity on vault tab entry**

Find the `else if (this.settingsTab === 'vault')` block and add:

```js
} else if (this.settingsTab === 'vault') {
  this.loadEmbedStatus();
  this.loadWorkers();
  this.loadPlasticity(); // ADD
}
```

**Step 4c: Add helper methods to app.js**

Add these methods adjacent to `loadWorkers()`:

```js
async loadPlasticity() {
  try {
    const data = await this.apiCall(
      '/api/admin/vault/' + encodeURIComponent(this.vault) + '/plasticity'
    );
    const cfg = data.config || {};
    this.plasticityForm.preset          = cfg.preset || 'default';
    this.plasticityForm.hebbianEnabled  = data.resolved?.hebbian_enabled ?? true;
    this.plasticityForm.decayEnabled    = data.resolved?.decay_enabled   ?? true;
    this.plasticityForm.hopDepth        = cfg.hop_depth       ?? null;
    this.plasticityForm.semanticWeight  = cfg.semantic_weight ?? null;
    this.plasticityForm.ftsWeight       = cfg.fts_weight      ?? null;
    this.plasticityForm.decayFloor      = cfg.decay_floor     ?? null;
    this.plasticityForm.decayStability  = cfg.decay_stability ?? null;
  } catch (_) {
    // Non-fatal — vault may not have explicit Plasticity config yet
  }
},

onPlasticityPresetChange() {
  // Reset advanced overrides when switching preset
  this.plasticityForm.hopDepth       = null;
  this.plasticityForm.semanticWeight = null;
  this.plasticityForm.ftsWeight      = null;
  this.plasticityForm.decayFloor     = null;
  this.plasticityForm.decayStability = null;
},

plasticityPresetDescription(preset) {
  const d = {
    'default':         'General-purpose. Decay on, Hebbian on, 2-hop BFS. Balanced weights.',
    'reference':       'Documentation and facts. Decay OFF — memories persist indefinitely.',
    'scratchpad':      'Ephemeral drafts. Aggressive decay (7-day stability, 0.01 floor). No Hebbian, no hops.',
    'knowledge-graph': 'Dense interlinked concepts. 4-hop BFS, slow decay (60-day stability).',
  };
  return d[preset] || '';
},

async savePlasticity() {
  this.plasticitySaving = true;
  this.plasticitySaveOk = false;
  this.plasticitySaveErr = '';
  try {
    const payload = { version: 1, preset: this.plasticityForm.preset };
    if (this.plasticityForm.showAdvanced) {
      if (this.plasticityForm.hopDepth       !== null) payload.hop_depth       = this.plasticityForm.hopDepth;
      if (this.plasticityForm.semanticWeight !== null) payload.semantic_weight = this.plasticityForm.semanticWeight;
      if (this.plasticityForm.ftsWeight      !== null) payload.fts_weight      = this.plasticityForm.ftsWeight;
      if (this.plasticityForm.decayFloor     !== null) payload.decay_floor     = this.plasticityForm.decayFloor;
      if (this.plasticityForm.decayStability !== null) payload.decay_stability = this.plasticityForm.decayStability;
      payload.hebbian_enabled = this.plasticityForm.hebbianEnabled;
      payload.decay_enabled   = this.plasticityForm.decayEnabled;
    }
    await this.apiCall(
      '/api/admin/vault/' + encodeURIComponent(this.vault) + '/plasticity',
      { method: 'PUT', body: JSON.stringify(payload) }
    );
    this.plasticitySaveOk = true;
    setTimeout(() => { this.plasticitySaveOk = false; }, 3000);
  } catch (err) {
    this.plasticitySaveErr = err.message;
    setTimeout(() => { this.plasticitySaveErr = ''; }, 5000);
  } finally {
    this.plasticitySaving = false;
  }
},
```

**Step 4d: Add Plasticity card to index.html**

In the vault settings section (after the Cognitive Engine Workers card, before the closing `</div>` of the vault tab), insert:

```html
<!-- ── Plasticity Configuration card ── -->
<div class="card-polished" style="max-width:560px;margin-top:1.5rem;">
  <h3 style="margin:0 0 0.25rem;font-size:1rem;">Plasticity</h3>
  <p style="color:var(--text-muted);font-size:0.8125rem;margin:0 0 1.25rem;">
    Controls how memories behave in this vault — decay rate, associative learning, and retrieval weights.
  </p>

  <!-- Preset selector -->
  <div class="form-group" style="margin-bottom:1rem;">
    <label>Preset</label>
    <select class="input-field" x-model="plasticityForm.preset" @change="onPlasticityPresetChange()">
      <option value="default">default — General (conversations, notes)</option>
      <option value="reference">reference — Documentation and facts (decay off)</option>
      <option value="scratchpad">scratchpad — Ephemeral notes and drafts (aggressive decay)</option>
      <option value="knowledge-graph">knowledge-graph — Dense interlinked concepts (4-hop BFS, slow decay)</option>
    </select>
  </div>

  <!-- Preset description -->
  <div style="background:var(--bg-elevated);border-radius:0.5rem;padding:0.625rem 0.875rem;margin-bottom:1rem;font-size:0.8125rem;color:var(--text-muted);"
       x-text="plasticityPresetDescription(plasticityForm.preset)"></div>

  <!-- Advanced toggle -->
  <div style="display:flex;align-items:center;gap:0.75rem;margin-bottom:1rem;">
    <label style="font-size:0.8125rem;font-weight:500;cursor:pointer;">
      <input type="checkbox" x-model="plasticityForm.showAdvanced" style="margin-right:0.4rem;" />
      Advanced overrides
    </label>
  </div>

  <!-- Advanced overrides panel -->
  <div x-show="plasticityForm.showAdvanced" style="border:1px solid var(--border);border-radius:0.5rem;padding:1rem;margin-bottom:1rem;">
    <p style="font-size:0.75rem;color:var(--text-muted);margin:0 0 1rem;">
      These override individual preset values. Leave blank to use the preset default.
    </p>

    <!-- Boolean toggles -->
    <div style="display:flex;gap:1.5rem;margin-bottom:1rem;flex-wrap:wrap;">
      <label style="font-size:0.8125rem;display:flex;align-items:center;gap:0.5rem;cursor:pointer;">
        <input type="checkbox" x-model="plasticityForm.hebbianEnabled" />
        Hebbian learning
      </label>
      <label style="font-size:0.8125rem;display:flex;align-items:center;gap:0.5rem;cursor:pointer;">
        <input type="checkbox" x-model="plasticityForm.decayEnabled" />
        Memory decay
      </label>
    </div>

    <!-- Numeric overrides -->
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:0.75rem;">
      <div class="form-group">
        <label style="font-size:0.75rem;">BFS Hop Depth (0–8)</label>
        <input class="input-field" type="number" min="0" max="8" x-model.number="plasticityForm.hopDepth" placeholder="preset" />
      </div>
      <div class="form-group">
        <label style="font-size:0.75rem;">Semantic Weight (0–1)</label>
        <input class="input-field" type="number" step="0.01" min="0" max="1" x-model.number="plasticityForm.semanticWeight" placeholder="preset" />
      </div>
      <div class="form-group">
        <label style="font-size:0.75rem;">FTS Weight (0–1)</label>
        <input class="input-field" type="number" step="0.01" min="0" max="1" x-model.number="plasticityForm.ftsWeight" placeholder="preset" />
      </div>
      <div class="form-group">
        <label style="font-size:0.75rem;">Decay Floor (0–1)</label>
        <input class="input-field" type="number" step="0.01" min="0" max="1" x-model.number="plasticityForm.decayFloor" placeholder="preset" />
      </div>
      <div class="form-group">
        <label style="font-size:0.75rem;">Decay Stability (days)</label>
        <input class="input-field" type="number" step="1" min="1" x-model.number="plasticityForm.decayStability" placeholder="preset" />
      </div>
    </div>
  </div>

  <!-- Save / status -->
  <div style="display:flex;align-items:center;gap:1rem;">
    <button class="btn-primary" @click="savePlasticity()" :disabled="plasticitySaving">
      <span x-text="plasticitySaving ? 'Saving…' : 'Save Plasticity'"></span>
    </button>
    <span x-show="plasticitySaveOk" style="color:var(--success);font-size:0.8125rem;">Saved</span>
    <span x-show="plasticitySaveErr" style="color:var(--danger);font-size:0.8125rem;" x-text="plasticitySaveErr"></span>
  </div>
</div>
```

**Step 4e: Build check**

```bash
go build ./...
```

Expected: no errors (HTML/JS changes don't affect Go build).

**Step 4f: Commit**

```bash
git add web/templates/index.html web/static/js/app.js
git commit -m "feat(plasticity): vault settings UI — Plasticity card with preset dropdown, advanced overrides, save/load"
```

---

## Task 5 — Final verification

**Step 5a: Run full test suite**

```bash
cd /Users/mjbonanno/github.com/scrypster/muninndb
go test ./... -timeout 120s
```

Expected: all pass.

**Step 5b: Build all binaries**

```bash
go build ./cmd/muninn/...
go build ./cmd/eval/...
go build ./cmd/bench/...
```

Expected: no errors.

---

## Sequencing Summary

| Task | Files | Notes |
|------|-------|-------|
| 1 | `internal/auth/plasticity.go`, `internal/auth/types.go` | Foundation — all others depend on this |
| 2 | `internal/transport/rest/admin_handlers.go`, `server.go` | REST layer |
| 3 | `internal/engine/engine.go`, all call sites | Engine integration + call site updates |
| 4 | `web/templates/index.html`, `web/static/js/app.js` | UI — independent of 2 and 3 |
| 5 | n/a | Final verification |

## Risk Notes

1. `activation.Weights` fields are `float64` in `internal/engine/activation/` — `resolved.SemanticWeight` is `float32`. Use `float64(resolved.SemanticWeight)` when building `actReq.Weights`.
2. The `hebbian_enabled` / `decay_enabled` JSON keys must match Go struct tags in `PlasticityConfig` exactly.
3. The admin `PUT` method must be in the CORS `Allow-Methods` header if not already — verify against existing PUT handlers in `server.go`.
4. In the UI, `null` numeric fields (when advanced is hidden) are not serialized — intentional.
5. Task 3 (engine) has 5 call sites for `NewEngine`. Verify all 5 compile before committing.
