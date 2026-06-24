package app

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/razad/razad/internal/crypto"
	"github.com/razad/razad/internal/database"
	"github.com/razad/razad/internal/process"
)

const (
	testUserID    = "user-1"
	testOrgID     = "org-1"
	testProjectID = "project-1"
	otherUserID   = "user-2"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	f := "/tmp/razad-app-test-" + t.Name() + ".db"
	os.Remove(f)
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close(); os.Remove(f) })
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	seedTenantData(t, db)
	return db
}

func seedUsersOnly(t *testing.T, db *sql.DB) {
	t.Helper()
	mustExec := func(q string, args ...any) {
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatalf("seed db: %v", err)
		}
	}
	mustExec(`INSERT INTO users (id, name, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`, testUserID, "Test User", "test@example.com", "hash")
	mustExec(`INSERT INTO users (id, name, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`, otherUserID, "Other User", "other@example.com", "hash")
}

func seedTenantData(t *testing.T, db *sql.DB) {
	t.Helper()
	seedUsersOnly(t, db)
	mustExec := func(q string, args ...any) {
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatalf("seed db: %v", err)
		}
	}
	mustExec(`INSERT INTO organizations (id, name, slug, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))`, testOrgID, "Test Org", "test-org")
	mustExec(`INSERT INTO organization_members (id, organization_id, user_id, role, created_at) VALUES (?, ?, ?, ?, datetime('now'))`, "member-1", testOrgID, testUserID, "admin")
	mustExec(`INSERT INTO projects (id, organization_id, name, slug, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`, testProjectID, testOrgID, "Test Project", "test-project")
}

func setupTestService(t *testing.T) *Service {
	t.Helper()
	db := setupTestDB(t)
	repo := NewRepository(db)
	dataDir := t.TempDir()
	proc := process.New(process.Config{Runner: "local", DataDir: dataDir})
	enc, _ := crypto.New("test-key-1234567890123456")
	return NewService(repo, proc, enc, dataDir)
}

func createGitRepoFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	repoDir := t.TempDir()
	for name, content := range files {
		fullPath := filepath.Join(repoDir, name)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir fixture path: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write fixture file %s: %v", name, err)
		}
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Razad Test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=Razad Test", "GIT_COMMITTER_EMAIL=test@example.com")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
		}
	}
	run("init")
	run("add", ".")
	run("commit", "-m", "fixture")
	return repoDir
}

func setupUserOnlyService(t *testing.T) *Service {
	t.Helper()
	f := "/tmp/razad-app-user-only-" + t.Name() + ".db"
	os.Remove(f)
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close(); os.Remove(f) })
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	seedUsersOnly(t, db)
	repo := NewRepository(db)
	proc := process.New(process.Config{Runner: "local", DataDir: t.TempDir()})
	enc, _ := crypto.New("test-key-1234567890123456")
	return NewService(repo, proc, enc, t.TempDir())
}

func TestCreateApp_Success(t *testing.T) {
	svc := setupTestService(t)

	app, err := svc.Create(testUserID, CreateAppRequest{
		Name:      "my-app",
		ProjectID: testProjectID,
		Runtime:   "node",
		StartCmd:  "npm start",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if app.Name != "my-app" {
		t.Errorf("expected name 'my-app', got %s", app.Name)
	}
	if app.Status != "created" {
		t.Errorf("expected status 'created', got %s", app.Status)
	}
}

func TestCreateApp_DeniedForForeignTenant(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.Create(otherUserID, CreateAppRequest{
		Name:      "my-app",
		ProjectID: testProjectID,
	})
	if err == nil {
		t.Fatal("expected error for foreign tenant")
	}
}

func TestCreateApp_BootstrapsDefaultWorkspaceForUserWithoutProject(t *testing.T) {
	svc := setupUserOnlyService(t)

	app, err := svc.Create(testUserID, CreateAppRequest{
		Name:      "bootstrap-app",
		ProjectID: "default",
		Runtime:   "node",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if app.ProjectID == "" {
		t.Fatal("expected project id to be assigned")
	}

	apps, err := svc.ListAll(testUserID)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 app after bootstrap, got %d", len(apps))
	}
	if apps[0].ID != app.ID {
		t.Fatalf("expected created app to be readable after bootstrap")
	}
}

func TestCreateApp_MissingName(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.Create(testUserID, CreateAppRequest{
		ProjectID: testProjectID,
	})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestGetApp_NotFound(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.Get(testUserID, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent app")
	}
}

func TestGetApp_Success(t *testing.T) {
	svc := setupTestService(t)

	created, _ := svc.Create(testUserID, CreateAppRequest{
		Name: "find-me", ProjectID: testProjectID, Runtime: "go",
	})

	app, err := svc.Get(testUserID, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if app.ID != created.ID {
		t.Errorf("expected app ID %s, got %s", created.ID, app.ID)
	}
}

func TestGetApp_DeniedForForeignTenant(t *testing.T) {
	svc := setupTestService(t)

	created, _ := svc.Create(testUserID, CreateAppRequest{
		Name: "private", ProjectID: testProjectID, Runtime: "go",
	})

	if _, err := svc.Get(otherUserID, created.ID); err == nil {
		t.Fatal("expected access denied")
	}
}

func TestListByProject_Empty(t *testing.T) {
	svc := setupTestService(t)

	apps, err := svc.ListByProject(testUserID, testProjectID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(apps) != 0 {
		t.Errorf("expected empty list, got %d", len(apps))
	}
}

func TestDeployAndStop(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	created, _ := svc.Create(testUserID, CreateAppRequest{
		Name: "deploy-test", ProjectID: testProjectID, Runtime: "node",
		StartCmd: "sleep 2",
	})

	// Deploy
	deployed, err := svc.Deploy(ctx, testUserID, created.ID)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	if deployed.Status != "running" {
		t.Errorf("expected status 'running', got %s", deployed.Status)
	}

	// Stop
	stopped, err := svc.Stop(ctx, testUserID, created.ID)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if stopped.Status != "stopped" {
		t.Errorf("expected status 'stopped', got %s", stopped.Status)
	}
}

type logStreamerRecorder struct {
	watched   []string
	unwatched []string
}

func (r *logStreamerRecorder) WatchApp(appID string) {
	r.watched = append(r.watched, appID)
}

func (r *logStreamerRecorder) UnwatchApp(appID string) {
	r.unwatched = append(r.unwatched, appID)
}

func TestDeleteApp(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	created, _ := svc.Create(testUserID, CreateAppRequest{
		Name: "delete-me", ProjectID: testProjectID, Runtime: "python",
	})

	if err := svc.Delete(ctx, testUserID, created.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := svc.Get(testUserID, created.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestDeployStartsLogStreamingAndDeleteStopsIt(t *testing.T) {
	svc := setupTestService(t)
	recorder := &logStreamerRecorder{}
	svc.SetLogStreamer(recorder)
	ctx := context.Background()

	created, _ := svc.Create(testUserID, CreateAppRequest{
		Name: "stream-me", ProjectID: testProjectID, Runtime: "node",
		StartCmd: "sleep 2",
	})

	if _, err := svc.Deploy(ctx, testUserID, created.ID); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if len(recorder.watched) != 1 || recorder.watched[0] != created.ID {
		t.Fatalf("expected WatchApp to be called with %s, got %#v", created.ID, recorder.watched)
	}

	if err := svc.Delete(ctx, testUserID, created.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if len(recorder.unwatched) != 1 || recorder.unwatched[0] != created.ID {
		t.Fatalf("expected UnwatchApp to be called with %s, got %#v", created.ID, recorder.unwatched)
	}
}

func TestSetEnvVars(t *testing.T) {
	svc := setupTestService(t)

	created, _ := svc.Create(testUserID, CreateAppRequest{
		Name: "env-test", ProjectID: testProjectID, Runtime: "node",
	})

	err := svc.SetEnvVars(testUserID, created.ID, []EnvVarInput{
		{Key: "DATABASE_URL", Value: "postgres://localhost"},
		{Key: "API_KEY", Value: "secret-123"},
	})
	if err != nil {
		t.Fatalf("SetEnvVars failed: %v", err)
	}

	vars, err := svc.GetEnvVarKeys(testUserID, created.ID)
	if err != nil {
		t.Fatalf("GetEnvVarKeys failed: %v", err)
	}

	if len(vars) != 2 {
		t.Errorf("expected 2 env vars, got %d", len(vars))
	}

	for _, v := range vars {
		if v.Value != "" {
			t.Errorf("expected empty value (masked), got %s", v.Value)
		}
	}
}

func TestListDeployments(t *testing.T) {
	svc := setupTestService(t)

	created, _ := svc.Create(testUserID, CreateAppRequest{
		Name: "deployments-test", ProjectID: testProjectID, Runtime: "node",
		StartCmd: "sleep 2",
	})

	ctx := context.Background()
	if _, err := svc.Deploy(ctx, testUserID, created.ID); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	deployments, err := svc.ListDeployments(testUserID, created.ID)
	if err != nil {
		t.Fatalf("ListDeployments failed: %v", err)
	}

	if len(deployments) == 0 {
		t.Fatal("expected at least one deployment")
	}
	if deployments[0].AppID != created.ID {
		t.Errorf("expected deployment app id %s, got %s", created.ID, deployments[0].AppID)
	}
	if deployments[0].Status != "success" {
		t.Errorf("expected deployment status success, got %s", deployments[0].Status)
	}
}

func TestDeploy_ClonesGitRepoIntoAppDir(t *testing.T) {
	svc := setupTestService(t)
	repoDir := createGitRepoFixture(t, map[string]string{
		"package.json": `{"name":"fixture-app","version":"1.0.0"}`,
		"README.md":    "fixture repo\n",
	})

	created, err := svc.Create(testUserID, CreateAppRequest{
		Name:      "git-clone-test",
		ProjectID: testProjectID,
		GitURL:    repoDir,
		Runtime:   "node",
		StartCmd:  "sleep 2",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if _, err := svc.Deploy(context.Background(), testUserID, created.ID); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	clonedFile := filepath.Join(svc.dataDir, "apps", created.ID, "package.json")
	if _, err := os.Stat(clonedFile); err != nil {
		t.Fatalf("expected cloned package.json at %s: %v", clonedFile, err)
	}
}

func TestDeploy_FailedStartMarksDeploymentFailed(t *testing.T) {
	svc := setupTestService(t)
	repoDir := createGitRepoFixture(t, map[string]string{
		"package.json": `{"name":"fixture-app","version":"1.0.0"}`,
	})

	created, err := svc.Create(testUserID, CreateAppRequest{
		Name:      "git-fail-test",
		ProjectID: testProjectID,
		GitURL:    repoDir,
		Runtime:   "node",
		StartCmd:  "sh -c 'exit 7'",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if _, err := svc.Deploy(context.Background(), testUserID, created.ID); err == nil {
		t.Fatal("expected deploy to fail for short-lived start command")
	}

	deployments, err := svc.ListDeployments(testUserID, created.ID)
	if err != nil {
		t.Fatalf("ListDeployments failed: %v", err)
	}
	if len(deployments) == 0 {
		t.Fatal("expected at least one deployment record")
	}
	if deployments[0].Status != "failed" {
		t.Fatalf("expected failed deployment status, got %s (log=%s)", deployments[0].Status, deployments[0].Log)
	}

	app, err := svc.Get(testUserID, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if app.Status != "failed" {
		t.Fatalf("expected app status failed, got %s", app.Status)
	}

	if !strings.Contains(deployments[0].Log, "start verification") {
		t.Fatalf("expected deployment log to mention verification failure, got: %s", deployments[0].Log)
	}
}

func TestDeploy_UsesRuntimeDetectionWhenStartCommandMissing(t *testing.T) {
	svc := setupTestService(t)
	repoDir := createGitRepoFixture(t, map[string]string{
		"package.json": `{"name":"fixture-app","version":"1.0.0","scripts":{"start":"node server.js"}}`,
		"server.js":    "setTimeout(() => process.exit(0), 1500)\n",
	})

	created, err := svc.Create(testUserID, CreateAppRequest{
		Name:      "git-detect-test",
		ProjectID: testProjectID,
		GitURL:    repoDir,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	deployed, err := svc.Deploy(context.Background(), testUserID, created.ID)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if deployed.Status != "running" {
		t.Fatalf("expected running app after detected runtime deploy, got %s", deployed.Status)
	}

	deployments, err := svc.ListDeployments(testUserID, created.ID)
	if err != nil {
		t.Fatalf("ListDeployments failed: %v", err)
	}
	if len(deployments) == 0 {
		t.Fatal("expected deployment history")
	}
	if !strings.Contains(deployments[0].Log, "Resolved runtime: node") {
		t.Fatalf("expected runtime detection log, got: %s", deployments[0].Log)
	}
	if !strings.Contains(deployments[0].Log, "Started: npm start") {
		t.Fatalf("expected detected npm start command, got: %s", deployments[0].Log)
	}
}

func TestDeploy_PassesConfiguredEnvVarsToProcess(t *testing.T) {
	svc := setupTestService(t)
	created, err := svc.Create(testUserID, CreateAppRequest{
		Name:      "env-pass-test",
		ProjectID: testProjectID,
		StartCmd:  `test "$APP_PORT" = "4123" && sleep 2`,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := svc.SetEnvVars(testUserID, created.ID, []EnvVarInput{{Key: "APP_PORT", Value: "4123"}}); err != nil {
		t.Fatalf("SetEnvVars failed: %v", err)
	}

	deployed, err := svc.Deploy(context.Background(), testUserID, created.ID)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if deployed.Status != "running" {
		t.Fatalf("expected running app after env deploy, got %s", deployed.Status)
	}
}
