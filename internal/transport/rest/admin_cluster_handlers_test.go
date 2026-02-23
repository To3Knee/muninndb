package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminCluster_GetToken_NoCoordinator(t *testing.T) {
	s := newTestServer(t, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/admin/cluster/token", nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminCluster_Enable_MissingRole(t *testing.T) {
	s := newTestServer(t, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/admin/cluster/enable", strings.NewReader(`{"bind_addr":"127.0.0.1:7777"}`))
	r.Header.Set("Content-Type", "application/json")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminCluster_Enable_MissingBindAddr(t *testing.T) {
	s := newTestServer(t, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/admin/cluster/enable", strings.NewReader(`{"role":"primary"}`))
	r.Header.Set("Content-Type", "application/json")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminCluster_Settings_Validation_HeartbeatNegative(t *testing.T) {
	s := newTestServer(t, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/admin/cluster/settings", strings.NewReader(`{"heartbeat_ms":-1}`))
	r.Header.Set("Content-Type", "application/json")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminCluster_Disable_NoCoordinator(t *testing.T) {
	s := newTestServer(t, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/admin/cluster/disable", nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
