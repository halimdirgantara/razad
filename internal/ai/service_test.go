package ai

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/razad/razad/internal/audit"
	db "github.com/razad/razad/internal/database"
)

const testAIUserID = "user-ai-1"

func setupAITestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := "/tmp/razad-ai-" + strings.ReplaceAll(t.Name(), "/", "_") + ".db"
	_ = os.Remove(path)
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
		_ = os.Remove(path)
	})
	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO users (id, name, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`, testAIUserID, "AI User", "ai@example.com", "hash"); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return conn
}

func TestServiceListsSupportedProvidersAndRegistry(t *testing.T) {
	svc := NewService(nil)
	caps := svc.Capabilities()

	if len(caps.Providers) < 4 {
		t.Fatalf("expected at least 4 providers, got %d", len(caps.Providers))
	}
	if !containsCapability(caps.AllowedActions, "restart_app") {
		t.Fatal("expected restart_app in allowed actions")
	}
	if !containsCapability(caps.BlockedActions, "delete_database") {
		t.Fatal("expected delete_database in blocked actions")
	}
}

func TestServiceRejectsBlockedAIAction(t *testing.T) {
	dbConn := setupAITestDB(t)
	aiSvc := NewService(audit.NewService(dbConn))

	_, err := aiSvc.RequestAction(context.Background(), testAIUserID, ActionRequest{Action: "delete_app", Target: "app-1"})
	if err == nil {
		t.Fatal("expected blocked action to fail")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("expected blocked error, got %v", err)
	}
}

func TestServiceRecordsAllowedAIActionInAudit(t *testing.T) {
	dbConn := setupAITestDB(t)
	auditSvc := audit.NewService(dbConn)
	aiSvc := NewService(auditSvc)

	result, err := aiSvc.RequestAction(context.Background(), testAIUserID, ActionRequest{Action: "restart_app", Target: "app-123", Reason: "crash loop"})
	if err != nil {
		t.Fatalf("RequestAction failed: %v", err)
	}
	if result.Status != "accepted" {
		t.Fatalf("expected accepted status, got %s", result.Status)
	}

	events, err := auditSvc.ListRecent(10)
	if err != nil {
		t.Fatalf("ListRecent failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(events))
	}
	if events[0].Action != "ai.action.requested" {
		t.Fatalf("expected ai.action.requested, got %s", events[0].Action)
	}
	if !strings.Contains(events[0].MetadataJSON, `"action":"restart_app"`) {
		t.Fatalf("expected action in metadata, got %s", events[0].MetadataJSON)
	}
}

func containsCapability(items []ActionCapability, name string) bool {
	for _, item := range items {
		if item.Name == name {
			return true
		}
	}
	return false
}
