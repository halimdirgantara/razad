package database

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/razad/razad/internal/auth"
)

const testUserID = "user-db-1"

func setupManagementTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := "/tmp/razad-db-mgmt-" + strings.ReplaceAll(t.Name(), "/", "_") + ".db"
	_ = os.Remove(path)
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
		_ = os.Remove(path)
	})
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO users (id, name, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`, testUserID, "DB User", "db@example.com", "hash"); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return db
}

func TestServiceCreateDatabaseInstanceGeneratesConnectionInfo(t *testing.T) {
	db := setupManagementTestDB(t)
	svc := NewService(NewRepository(db))

	inst, err := svc.Create(testUserID, CreateRequest{
		Name:   "Primary Postgres",
		Engine: "postgresql",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if inst.OwnerUserID != testUserID {
		t.Fatalf("expected owner %s, got %s", testUserID, inst.OwnerUserID)
	}
	if inst.Engine != "postgresql" {
		t.Fatalf("expected engine postgresql, got %s", inst.Engine)
	}
	if inst.Status != "provisioned" {
		t.Fatalf("expected status provisioned, got %s", inst.Status)
	}
	if inst.Username == "" || inst.Password == "" {
		t.Fatal("expected generated credentials")
	}
	if !strings.Contains(inst.ConnectionString, inst.Host) || !strings.Contains(inst.ConnectionString, inst.DatabaseName) {
		t.Fatalf("expected connection string to include host and database name, got %s", inst.ConnectionString)
	}
}

func TestServiceListReturnsCreatedDatabaseInstances(t *testing.T) {
	db := setupManagementTestDB(t)
	svc := NewService(NewRepository(db))

	created, err := svc.Create(testUserID, CreateRequest{Name: "Redis Cache", Engine: "redis"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	instances, err := svc.List(testUserID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].ID != created.ID {
		t.Fatalf("expected %s, got %s", created.ID, instances[0].ID)
	}
}

func TestHandlerCreateAndListDatabaseInstances(t *testing.T) {
	db := setupManagementTestDB(t)
	svc := NewService(NewRepository(db))
	h := NewHandler(svc)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/databases", strings.NewReader(`{"name":"MySQL Primary","engine":"mysql"}`))
	createReq = createReq.WithContext(context.WithValue(createReq.Context(), auth.ContextUserIDKey, testUserID))
	createRR := httptest.NewRecorder()

	h.Create(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createRR.Code)
	}
	if !strings.Contains(createRR.Body.String(), `"engine":"mysql"`) {
		t.Fatalf("expected mysql response, got %s", createRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/databases", nil)
	listReq = listReq.WithContext(context.WithValue(listReq.Context(), auth.ContextUserIDKey, testUserID))
	listRR := httptest.NewRecorder()

	h.List(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listRR.Code)
	}
	if !strings.Contains(listRR.Body.String(), `"name":"MySQL Primary"`) {
		t.Fatalf("expected created instance in list response, got %s", listRR.Body.String())
	}
}
