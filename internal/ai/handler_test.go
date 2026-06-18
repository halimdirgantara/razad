package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/razad/razad/internal/auth"
)

func TestHandlerIndexReturnsCapabilities(t *testing.T) {
	h := NewHandler(NewService(nil))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai", nil)
	rr := httptest.NewRecorder()

	h.Index(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"restart_app"`) {
		t.Fatalf("expected allowed action in response, got %s", rr.Body.String())
	}
}

func TestHandlerActionRequiresAuth(t *testing.T) {
	h := NewHandler(NewService(nil))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/actions", strings.NewReader(`{"action":"restart_app"}`))
	rr := httptest.NewRecorder()

	h.Action(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandlerActionRejectsBlockedAction(t *testing.T) {
	h := NewHandler(NewService(nil))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/actions", strings.NewReader(`{"action":"delete_app"}`))
	req = req.WithContext(context.WithValue(req.Context(), auth.ContextUserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.Action(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}
