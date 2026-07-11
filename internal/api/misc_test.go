package api_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/config"
	"github.com/razad/razad/internal/requestctx"
)

// setupMiscDB opens an in-memory-ish sqlite db with just the tables the
// misc handler tests need. We can't import internal/database.Migrate
// here because internal/database imports internal/api (for WriteError /
// WriteJSON used by database/handler.go) — that would be a test-binary
// import cycle. Keeping the schema inline keeps these tests isolated and
// cycle-free.
func setupMiscDB(t *testing.T) (*sql.DB, *config.Config) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "misc.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	stmts := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY, organization_id TEXT NOT NULL, name TEXT NOT NULL, slug TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS apps (
			id TEXT PRIMARY KEY, project_id TEXT NOT NULL, name TEXT NOT NULL,
			git_url TEXT, runtime TEXT NOT NULL DEFAULT 'unknown', start_cmd TEXT,
			status TEXT NOT NULL DEFAULT 'created',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS app_deployments (
			id TEXT PRIMARY KEY, app_id TEXT NOT NULL, version TEXT NOT NULL DEFAULT 'latest',
			status TEXT NOT NULL DEFAULT 'pending', log TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("setup: %v (stmt=%s)", err, s)
		}
	}
	cfg := config.Defaults()
	return db, &cfg
}

func TestMiscServicesRequiresAuth(t *testing.T) {
	db, cfg := setupMiscDB(t)
	h := api.NewMiscHandler(db, *cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	rr := httptest.NewRecorder()
	h.Services(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMiscServicesReturnsEmptyList(t *testing.T) {
	db, cfg := setupMiscDB(t)
	h := api.NewMiscHandler(db, *cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	req = req.WithContext(requestctx.WithUserID(req.Context(), "u-1"))
	rr := httptest.NewRecorder()
	h.Services(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		Items []map[string]any `json:"items"`
		Count int              `json:"count"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Always includes the daemon entry.
	if resp.Count != 1 {
		t.Errorf("count: got %d, want 1 (just the daemon)", resp.Count)
	}
	if name, _ := resp.Items[0]["name"].(string); name != "razad-daemon" {
		t.Errorf("first item name: got %q, want razad-daemon", name)
	}
}

func TestMiscSettingsRedactsSecretKey(t *testing.T) {
	_, cfg := setupMiscDB(t)
	cfg.Auth.SecretKey = "super-secret-do-not-leak"
	h := api.NewMiscHandler(nil, *cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	req = req.WithContext(requestctx.WithUserID(req.Context(), "u-1"))
	rr := httptest.NewRecorder()
	h.Settings(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if contains(body, "super-secret-do-not-leak") {
		t.Error("response leaked the secret key")
	}
	if !contains(body, "[redacted]") {
		t.Errorf("expected [redacted] in response, got: %s", body)
	}
}

func TestMiscSettingsRequiresAuth(t *testing.T) {
	_, cfg := setupMiscDB(t)
	h := api.NewMiscHandler(nil, *cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	rr := httptest.NewRecorder()
	h.Settings(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMiscDeploymentsRequiresAuth(t *testing.T) {
	db, cfg := setupMiscDB(t)
	h := api.NewMiscHandler(db, *cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/deployments", nil)
	rr := httptest.NewRecorder()
	h.Deployments(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMiscDeploymentsReturnsEmptyList(t *testing.T) {
	db, cfg := setupMiscDB(t)
	h := api.NewMiscHandler(db, *cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/deployments", nil)
	req = req.WithContext(requestctx.WithUserID(req.Context(), "u-1"))
	rr := httptest.NewRecorder()
	h.Deployments(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Items []map[string]any `json:"items"`
		Count int              `json:"count"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("expected empty list, got %d", resp.Count)
	}
}

func TestMiscServicesRejectsNonGET(t *testing.T) {
	db, cfg := setupMiscDB(t)
	h := api.NewMiscHandler(db, *cfg)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/services", nil)
	req = req.WithContext(requestctx.WithUserID(req.Context(), "u-1"))
	rr := httptest.NewRecorder()
	h.Services(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
