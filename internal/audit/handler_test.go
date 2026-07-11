package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/razad/razad/internal/database"
	"github.com/razad/razad/internal/requestctx"
)

func setupHandlerDB(t *testing.T) {
	t.Helper()
	file := "/tmp/razad-audit-handler-" + t.Name() + ".db"
	os.Remove(file)
	db, err := sqlOpen(file)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO users (id, name, email, password_hash) VALUES (?, ?, ?, ?)`,
		"user-1", "Test User", "user-1@example.com", "x",
	); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(file)
	})
}

func TestListRequiresAuth(t *testing.T) {
	setupHandlerDB(t)
	db, _ := sqlOpen("/tmp/razad-audit-handler-TestListRequiresAuth.db")
	h := NewHandler(NewService(db))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestListReturnsPageWithTotal(t *testing.T) {
	setupHandlerDB(t)
	db, _ := sqlOpen("/tmp/razad-audit-handler-TestListReturnsPageWithTotal.db")
	svc := NewService(db)
	h := NewHandler(svc)
	ctx := context.Background()

	// Insert three events.
	for _, action := range []string{"app.create", "app.deploy", "app.stop"} {
		if err := svc.Record(ctx, "user-1", action, "app", "app-1", map[string]any{"action": action}); err != nil {
			t.Fatalf("record %s: %v", action, err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit?limit=2&offset=0", nil)
	req = req.WithContext(requestctx.WithUserID(req.Context(), "user-1"))
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Items  []Event `json:"items"`
		Total  int     `json:"total"`
		Limit  int     `json:"limit"`
		Offset int     `json:"offset"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total != 3 {
		t.Errorf("total: got %d, want 3", resp.Total)
	}
	if len(resp.Items) != 2 {
		t.Errorf("items: got %d, want 2 (limit)", len(resp.Items))
	}
	if resp.Limit != 2 {
		t.Errorf("limit: got %d, want 2", resp.Limit)
	}
	if resp.Offset != 0 {
		t.Errorf("offset: got %d, want 0", resp.Offset)
	}
	// Don't assert a specific first-item order: all three events were inserted
	// in the same second so the (created_at DESC, id DESC) tiebreaker reduces
	// to random ID order, which is fine for pagination correctness but not
	// deterministic enough for an exact-match assertion. Verify the page
	// contains the expected actions instead.
	wantActions := map[string]bool{"app.create": true, "app.deploy": true, "app.stop": true}
	gotActions := map[string]bool{}
	for _, e := range resp.Items {
		gotActions[e.Action] = true
	}
	for a := range wantActions {
		if !gotActions[a] {
			// Not in first page — try the second.
			// (handled by re-querying with offset=2 below)
		}
	}
	// Across two pages we must see all three actions exactly once each.
	page1 := resp.Items
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/audit?limit=2&offset=2", nil)
	req2 = req2.WithContext(requestctx.WithUserID(req2.Context(), "user-1"))
	rr2 := httptest.NewRecorder()
	h.List(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("page2: %d body=%s", rr2.Code, rr2.Body.String())
	}
	var resp2 struct {
		Items  []Event `json:"items"`
		Total  int     `json:"total"`
		Limit  int     `json:"limit"`
		Offset int     `json:"offset"`
	}
	if err := json.NewDecoder(rr2.Body).Decode(&resp2); err != nil {
		t.Fatalf("decode page2: %v", err)
	}
	if resp2.Total != 3 || resp2.Offset != 2 {
		t.Errorf("page2 metadata: total=%d offset=%d", resp2.Total, resp2.Offset)
	}
	seen := map[string]int{}
	for _, e := range page1 {
		seen[e.Action]++
	}
	for _, e := range resp2.Items {
		seen[e.Action]++
	}
	for a := range wantActions {
		if seen[a] != 1 {
			t.Errorf("action %s appeared %d times across both pages, want 1", a, seen[a])
		}
	}
}

func TestListRejectsNonGET(t *testing.T) {
	setupHandlerDB(t)
	db, _ := sqlOpen("/tmp/razad-audit-handler-TestListRejectsNonGET.db")
	h := NewHandler(NewService(db))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/audit", nil)
	req = req.WithContext(requestctx.WithUserID(req.Context(), "user-1"))
	rr := httptest.NewRecorder()
	h.List(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
