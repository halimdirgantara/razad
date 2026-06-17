package audit

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/razad/razad/internal/database"
)

func setupAuditDB(t *testing.T) *sql.DB {
	t.Helper()
	file := "/tmp/razad-audit-test-" + t.Name() + ".db"
	os.Remove(file)
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close(); os.Remove(file) })
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestRecordAndListRecent(t *testing.T) {
	db := setupAuditDB(t)
	svc := NewService(db)
	if err := svc.Record(context.Background(), "user-1", "app.create", "app", "app-1", map[string]any{"name": "hello"}); err != nil {
		t.Fatalf("record: %v", err)
	}
	events, err := svc.ListRecent(10)
	if err != nil {
		t.Fatalf("list recent: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Action != "app.create" {
		t.Fatalf("unexpected action: %s", events[0].Action)
	}
}
