package database

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/razad/razad/internal/auth"
	"github.com/razad/razad/internal/process"
)

const testUserID = "user-db-1"

type fakeRunner struct {
	startCalls   []startCall
	stopCalls    []string
	restartCalls []string
	status       process.ProcessState
}

type startCall struct {
	name    string
	command string
	env     []string
	workDir string
}

func (f *fakeRunner) Start(ctx context.Context, name, command string, env []string, workDir string) error {
	f.startCalls = append(f.startCalls, startCall{name: name, command: command, env: env, workDir: workDir})
	return nil
}

func (f *fakeRunner) Stop(ctx context.Context, name string) error {
	f.stopCalls = append(f.stopCalls, name)
	return nil
}

func (f *fakeRunner) Restart(ctx context.Context, name string) error {
	f.restartCalls = append(f.restartCalls, name)
	return nil
}

func (f *fakeRunner) Status(ctx context.Context, name string) (process.ProcessState, error) {
	return f.status, nil
}

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

func TestServiceCreateProvisionsDatabaseService(t *testing.T) {
	db := setupManagementTestDB(t)
	runner := &fakeRunner{}
	svc := NewService(NewRepository(db), runner, t.TempDir())

	inst, err := svc.Create(testUserID, CreateRequest{
		Name:   "Primary Postgres",
		Engine: "postgresql",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if inst.Status != "running" {
		t.Fatalf("expected status running after provisioning, got %s", inst.Status)
	}
	if len(runner.startCalls) != 1 {
		t.Fatalf("expected 1 start call, got %d", len(runner.startCalls))
	}
	call := runner.startCalls[0]
	if call.name != inst.ID {
		t.Fatalf("expected start name %s, got %s", inst.ID, call.name)
	}
	if !strings.Contains(call.command, "postgres") {
		t.Fatalf("expected postgres command, got %s", call.command)
	}
	if !strings.Contains(call.command, "-D") {
		t.Fatalf("expected datadir flag in command, got %s", call.command)
	}
	if !strings.Contains(call.workDir, filepath.Join("databases", inst.ID)) {
		t.Fatalf("expected workdir under databases dir, got %s", call.workDir)
	}
	if !strings.Contains(inst.ConnectionString, inst.Host) || !strings.Contains(inst.ConnectionString, inst.DatabaseName) {
		t.Fatalf("expected connection string to include host and database name, got %s", inst.ConnectionString)
	}
}

func TestServiceDeleteStopsProvisionedDatabaseService(t *testing.T) {
	db := setupManagementTestDB(t)
	runner := &fakeRunner{}
	svc := NewService(NewRepository(db), runner, t.TempDir())

	inst, err := svc.Create(testUserID, CreateRequest{Name: "Redis Cache", Engine: "redis"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := svc.Delete(testUserID, inst.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if len(runner.stopCalls) != 1 || runner.stopCalls[0] != inst.ID {
		t.Fatalf("expected stop call for %s, got %#v", inst.ID, runner.stopCalls)
	}
	if _, err := svc.Get(testUserID, inst.ID); err == nil {
		t.Fatal("expected deleted instance to be gone")
	}
}

func TestServiceStatusAndRestartUseRunnerState(t *testing.T) {
	db := setupManagementTestDB(t)
	runner := &fakeRunner{status: process.StateRunning}
	svc := NewService(NewRepository(db), runner, t.TempDir())

	inst, err := svc.Create(testUserID, CreateRequest{Name: "Mongo Main", Engine: "mongodb"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	runner.status = process.StateStopped
	refreshed, err := svc.Status(testUserID, inst.ID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if refreshed.Status != string(process.StateStopped) {
		t.Fatalf("expected stopped status, got %s", refreshed.Status)
	}

	if _, err := svc.Restart(testUserID, inst.ID); err != nil {
		t.Fatalf("Restart failed: %v", err)
	}
	if len(runner.startCalls) < 2 {
		t.Fatalf("expected restart to start service again, got %#v", runner.startCalls)
	}
}

func TestServiceListReturnsCreatedDatabaseInstances(t *testing.T) {
	db := setupManagementTestDB(t)
	svc := NewService(NewRepository(db), &fakeRunner{}, t.TempDir())

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
	svc := NewService(NewRepository(db), &fakeRunner{}, t.TempDir())
	h := NewHandler(svc, nil)

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
