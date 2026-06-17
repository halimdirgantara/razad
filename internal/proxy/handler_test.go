package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/razad/razad/internal/auth"
)

func TestHandlerRender(t *testing.T) {
	h := NewHandler(NewService(t.TempDir()), nil)
	body, _ := json.Marshal(map[string]any{
		"name":          "app-one",
		"domain":        "app.example.com",
		"upstream_host": "127.0.0.1",
		"upstream_port": 8080,
		"tls":           false,
		"body_limit_mb": 20,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/proxy/render", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.ContextUserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.Render(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "proxy_pass http://127.0.0.1:8080;") {
		t.Fatalf("response missing config: %s", rr.Body.String())
	}
}

func TestHandlerApplyAndRollback(t *testing.T) {
	base := t.TempDir()
	h := NewHandler(NewService(base), nil)
	binding := map[string]any{
		"name":          "app-one",
		"domain":        "app.example.com",
		"upstream_host": "127.0.0.1",
		"upstream_port": 8080,
	}

	applyBody, _ := json.Marshal(binding)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/proxy/apply", bytes.NewReader(applyBody))
	req = req.WithContext(context.WithValue(req.Context(), auth.ContextUserIDKey, "user-1"))
	rr := httptest.NewRecorder()
	h.Apply(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("apply failed: %d %s", rr.Code, rr.Body.String())
	}
	for _, p := range []string{"sites-available/app-one.conf", "sites-enabled/app-one.conf", "backups/app-one.conf.bak"} {
		if _, err := os.Stat(filepath.Join(base, p)); err != nil {
			t.Fatalf("expected file %s: %v", p, err)
		}
	}

	// Simulate drift and rollback.
	enabled := filepath.Join(base, "sites-enabled", "app-one.conf")
	if err := os.WriteFile(enabled, []byte("changed"), 0o644); err != nil {
		t.Fatalf("write drift failed: %v", err)
	}
	rbReq := httptest.NewRequest(http.MethodPost, "/api/v1/proxy/rollback", bytes.NewReader(applyBody))
	rbReq = rbReq.WithContext(context.WithValue(rbReq.Context(), auth.ContextUserIDKey, "user-1"))
	rr = httptest.NewRecorder()
	h.Rollback(rr, rbReq)
	if rr.Code != http.StatusOK {
		t.Fatalf("rollback failed: %d %s", rr.Code, rr.Body.String())
	}
	data, err := os.ReadFile(enabled)
	if err != nil {
		t.Fatalf("read enabled failed: %v", err)
	}
	if !strings.Contains(string(data), "server_name app.example.com;") {
		t.Fatalf("rollback did not restore config: %s", string(data))
	}
}
