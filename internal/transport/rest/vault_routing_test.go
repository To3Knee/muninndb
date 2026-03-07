package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/scrypster/muninndb/internal/auth"
)

// vaultTrackingEngine wraps MockEngine and records the vault passed to key engine calls.
type vaultTrackingEngine struct {
	MockEngine
	lastWriteVault    string
	lastActivateVault string
	lastListVault     string
	lastReadVault     string
	lastForgetVault   string
}

func (e *vaultTrackingEngine) Write(ctx context.Context, req *WriteRequest) (*WriteResponse, error) {
	e.lastWriteVault = req.Vault
	return e.MockEngine.Write(ctx, req)
}

func (e *vaultTrackingEngine) Activate(ctx context.Context, req *ActivateRequest) (*ActivateResponse, error) {
	e.lastActivateVault = req.Vault
	return e.MockEngine.Activate(ctx, req)
}

func (e *vaultTrackingEngine) ListEngrams(ctx context.Context, req *ListEngramsRequest) (*ListEngramsResponse, error) {
	e.lastListVault = req.Vault
	return e.MockEngine.ListEngrams(ctx, req)
}

func (e *vaultTrackingEngine) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	e.lastReadVault = req.Vault
	return e.MockEngine.Read(ctx, req)
}

func (e *vaultTrackingEngine) Forget(ctx context.Context, req *ForgetRequest) (*ForgetResponse, error) {
	e.lastForgetVault = req.Vault
	return e.MockEngine.Forget(ctx, req)
}

// newVaultTrackingServer creates a Server with a vaultTrackingEngine and a
// public "default" vault. The store is returned so tests can configure auth.
func newVaultTrackingServer(t *testing.T) (*Server, *vaultTrackingEngine, *auth.Store) {
	t.Helper()
	eng := &vaultTrackingEngine{}
	store := newTestAuthStore(t)
	if err := store.SetVaultConfig(auth.VaultConfig{Name: "default", Public: true}); err != nil {
		t.Fatalf("SetVaultConfig: %v", err)
	}
	srv := NewServer("localhost:0", eng, store, nil, nil, EmbedInfo{}, EnrichInfo{}, nil, "", nil)
	return srv, eng, store
}

// TestVaultRouting_Write_DefaultVault verifies that POST /api/engrams with no
// vault param passes "default" to the engine.
func TestVaultRouting_Write_DefaultVault(t *testing.T) {
	srv, eng, _ := newVaultTrackingServer(t)

	body := strings.NewReader(`{"concept":"test","content":"hello"}`)
	req := httptest.NewRequest("POST", "/api/engrams", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if eng.lastWriteVault != "default" {
		t.Errorf("engine Write vault: want %q, got %q", "default", eng.lastWriteVault)
	}
}

// TestVaultRouting_Write_ExplicitVault verifies that POST /api/engrams?vault=myvault
// passes "myvault" to the engine.
func TestVaultRouting_Write_ExplicitVault(t *testing.T) {
	srv, eng, store := newVaultTrackingServer(t)
	if err := store.SetVaultConfig(auth.VaultConfig{Name: "myvault", Public: true}); err != nil {
		t.Fatalf("SetVaultConfig: %v", err)
	}

	body := strings.NewReader(`{"concept":"test","content":"hello"}`)
	req := httptest.NewRequest("POST", "/api/engrams?vault=myvault", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if eng.lastWriteVault != "myvault" {
		t.Errorf("engine Write vault: want %q, got %q", "myvault", eng.lastWriteVault)
	}
}

// TestVaultRouting_Activate_ExplicitVault verifies that POST /api/activate?vault=myvault
// passes "myvault" to the engine.
func TestVaultRouting_Activate_ExplicitVault(t *testing.T) {
	srv, eng, store := newVaultTrackingServer(t)
	if err := store.SetVaultConfig(auth.VaultConfig{Name: "myvault", Public: true}); err != nil {
		t.Fatalf("SetVaultConfig: %v", err)
	}

	body := strings.NewReader(`{"context":["something"]}`)
	req := httptest.NewRequest("POST", "/api/activate?vault=myvault", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if eng.lastActivateVault != "myvault" {
		t.Errorf("engine Activate vault: want %q, got %q", "myvault", eng.lastActivateVault)
	}
}

// TestVaultRouting_ListEngrams_ExplicitVault verifies that GET /api/engrams?vault=myvault
// passes "myvault" to the engine.
func TestVaultRouting_ListEngrams_ExplicitVault(t *testing.T) {
	srv, eng, store := newVaultTrackingServer(t)
	if err := store.SetVaultConfig(auth.VaultConfig{Name: "myvault", Public: true}); err != nil {
		t.Fatalf("SetVaultConfig: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/engrams?vault=myvault", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if eng.lastListVault != "myvault" {
		t.Errorf("engine ListEngrams vault: want %q, got %q", "myvault", eng.lastListVault)
	}
}

// TestVaultAuth_LockedVaultRejectedAtEndpoint verifies that a locked vault
// rejects unauthenticated requests with 401 at the endpoint level.
func TestVaultAuth_LockedVaultRejectedAtEndpoint(t *testing.T) {
	srv, _, store := newVaultTrackingServer(t)
	if err := store.SetVaultConfig(auth.VaultConfig{Name: "locked", Public: false}); err != nil {
		t.Fatalf("SetVaultConfig: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/engrams?vault=locked", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("locked vault no key: want 401, got %d", w.Code)
	}
}

// TestVaultAuth_ValidKeyGrantsAccess verifies that a valid scoped API key
// passes auth and reaches the engine with the correct vault.
func TestVaultAuth_ValidKeyGrantsAccess(t *testing.T) {
	srv, eng, store := newVaultTrackingServer(t)
	if err := store.SetVaultConfig(auth.VaultConfig{Name: "secured", Public: false}); err != nil {
		t.Fatalf("SetVaultConfig: %v", err)
	}
	token, _, err := store.GenerateAPIKey("secured", "agent", "full", nil)
	if err != nil {
		t.Fatalf("GenerateAPIKey: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/engrams?vault=secured", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("valid key: want 200, got %d: %s", w.Code, w.Body.String())
	}
	if eng.lastListVault != "secured" {
		t.Errorf("engine vault: want %q, got %q", "secured", eng.lastListVault)
	}
}

// TestVaultAuth_KeyMismatchRejected verifies that a key scoped to vault-a
// cannot access vault-b, even through the full endpoint path.
func TestVaultAuth_KeyMismatchRejected(t *testing.T) {
	srv, _, store := newVaultTrackingServer(t)
	if err := store.SetVaultConfig(auth.VaultConfig{Name: "vault-a", Public: false}); err != nil {
		t.Fatalf("SetVaultConfig vault-a: %v", err)
	}
	if err := store.SetVaultConfig(auth.VaultConfig{Name: "vault-b", Public: false}); err != nil {
		t.Fatalf("SetVaultConfig vault-b: %v", err)
	}
	token, _, err := store.GenerateAPIKey("vault-a", "agent", "full", nil)
	if err != nil {
		t.Fatalf("GenerateAPIKey: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/engrams?vault=vault-b", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("key mismatch: want 401, got %d", w.Code)
	}
}

// TestVaultRouting_Read_ExplicitVault verifies that GET /api/engrams/{id}?vault=myvault
// passes "myvault" to the engine.
func TestVaultRouting_Read_ExplicitVault(t *testing.T) {
	srv, eng, store := newVaultTrackingServer(t)
	if err := store.SetVaultConfig(auth.VaultConfig{Name: "myvault", Public: true}); err != nil {
		t.Fatalf("SetVaultConfig: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/engrams/some-id?vault=myvault", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	// MockEngine.Read returns a valid ReadResponse with nil error; 200 is expected.
	// We care that the vault was correctly forwarded, not the HTTP status.
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if eng.lastReadVault != "myvault" {
		t.Errorf("engine Read vault: want %q, got %q", "myvault", eng.lastReadVault)
	}
}

// TestVaultRouting_Forget_ExplicitVault verifies that DELETE /api/engrams/{id}?vault=myvault
// passes "myvault" to the engine.
func TestVaultRouting_Forget_ExplicitVault(t *testing.T) {
	srv, eng, store := newVaultTrackingServer(t)
	if err := store.SetVaultConfig(auth.VaultConfig{Name: "myvault", Public: true}); err != nil {
		t.Fatalf("SetVaultConfig: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/engrams/some-id?vault=myvault", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	// MockEngine.Forget returns &ForgetResponse{OK: true} with nil error; check vault was forwarded.
	if eng.lastForgetVault != "myvault" {
		t.Errorf("engine Forget vault: want %q, got %q", "myvault", eng.lastForgetVault)
	}
}
