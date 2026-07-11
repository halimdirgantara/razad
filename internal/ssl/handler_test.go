package ssl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/razad/razad/internal/auth"
)

func TestHandlerIssue(t *testing.T) {
	h := NewHandler(NewService(t.TempDir()), nil, nil)
	body, _ := json.Marshal(map[string]any{
		"domain":  "app.example.com",
		"email":   "ops@example.com",
		"webroot": "/var/www/html",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ssl/issue", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.ContextUserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.Issue(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "certbot certonly") {
		t.Fatalf("response missing certbot command: %s", rr.Body.String())
	}
}

func TestHandlerRenew(t *testing.T) {
	h := NewHandler(NewService(t.TempDir()), nil, nil)
	body, _ := json.Marshal(map[string]any{"domain": "app.example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ssl/renew", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.ContextUserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.Renew(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "certbot renew --cert-name app.example.com") {
		t.Fatalf("response missing renew command: %s", rr.Body.String())
	}
}
