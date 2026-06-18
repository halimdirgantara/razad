package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razad/razad/internal/auth"
)

func TestDeploymentsHandlerReturnsRecentDeployments(t *testing.T) {
	svc := setupTestService(t)
	h := NewHandler(svc)

	created, err := svc.Create(testUserID, CreateAppRequest{
		Name: "detail-test",
		ProjectID: testProjectID,
		Runtime: "node",
		StartCmd: "echo hello",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if _, err := svc.Deploy(context.Background(), testUserID, created.ID); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+created.ID+"/deployments", nil)
	req = req.WithContext(context.WithValue(req.Context(), auth.ContextUserIDKey, testUserID))
	rr := httptest.NewRecorder()

	h.Deployments(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body []map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected deployments in response")
	}
	if body[0]["app_id"] != created.ID {
		t.Fatalf("expected app_id %s, got %v", created.ID, body[0]["app_id"])
	}
}
