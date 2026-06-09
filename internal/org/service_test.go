package org

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/razad/razad/internal/database"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile := "/tmp/razad-org-test-" + t.Name() + ".db"
	os.Remove(tmpFile)

	db, err := sql.Open("sqlite3", tmpFile)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile)
	})

	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	db := setupTestDB(t)
	repo := NewRepository(db)
	return NewService(repo)
}

const testUserID = "test-user-001"

func TestCreate_Success(t *testing.T) {
	svc := newTestService(t)

	org, err := svc.Create("My Org", "my-org", testUserID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if org.Name != "My Org" {
		t.Errorf("expected name 'My Org', got %s", org.Name)
	}
	if org.Slug != "my-org" {
		t.Errorf("expected slug 'my-org', got %s", org.Slug)
	}
}

func TestCreate_InvalidSlug(t *testing.T) {
	svc := newTestService(t)

	tests := []string{"", "ab", "has spaces", "a-"}
	for _, slug := range tests {
		_, err := svc.Create("Test", slug, testUserID)
		if err != ErrInvalidSlug {
			t.Errorf("expected ErrInvalidSlug for slug %q, got %v", slug, err)
		}
	}
}

func TestCreate_DuplicateSlug(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.Create("First", "dup-slug", testUserID)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = svc.Create("Second", "dup-slug", testUserID)
	if err != ErrSlugTaken {
		t.Errorf("expected ErrSlugTaken, got %v", err)
	}
}

func TestList_Empty(t *testing.T) {
	svc := newTestService(t)

	orgs, err := svc.List("nonexistent-user")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(orgs) != 0 {
		t.Errorf("expected empty list, got %d orgs", len(orgs))
	}
}

func TestList_WithOrgs(t *testing.T) {
	svc := newTestService(t)

	svc.Create("Org Alpha", "alpha", testUserID)
	svc.Create("Org Beta", "beta", testUserID)

	orgs, err := svc.List(testUserID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(orgs) != 2 {
		t.Errorf("expected 2 orgs, got %d", len(orgs))
	}
}

func TestGet_NotFound(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.Get("nonexistent-id", testUserID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGet_Success(t *testing.T) {
	svc := newTestService(t)

	created, _ := svc.Create("Findable Org", "findable", testUserID)

	org, err := svc.Get(created.ID, testUserID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if org.ID != created.ID {
		t.Errorf("expected org ID %s, got %s", created.ID, org.ID)
	}
}

func TestAddMember_NotAdmin(t *testing.T) {
	svc := newTestService(t)

	org, _ := svc.Create("Non-member test", "non-member", testUserID)

	err := svc.AddMember(org.ID, "other-user", "new-user", "member")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound for non-member actor, got %v", err)
	}
}
