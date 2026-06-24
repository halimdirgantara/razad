package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/razad/razad/internal/audit"
	"github.com/razad/razad/internal/crypto"
	"github.com/razad/razad/internal/domain"
	"github.com/razad/razad/internal/process"
	runtimepkg "github.com/razad/razad/internal/runtime"
)

var (
	ErrNotFound           = errors.New("app: not found")
	ErrNotRunning         = errors.New("app: not running")
	ErrAccessDenied       = errors.New("app: access denied")
	ErrProjectUnavailable = errors.New("app: no accessible project found for user")
)

type LogStreamer interface {
	WatchApp(appID string)
	UnwatchApp(appID string)
}

// Service handles app management business logic.
type Service struct {
	repo       *Repository
	proc       process.Runner
	enc        *crypto.Encrypter
	audit      *audit.Service
	dataDir    string
	logStreams LogStreamer
}

// NewService creates an app service.
func NewService(repo *Repository, proc process.Runner, enc *crypto.Encrypter, dataDir string) *Service {
	return &Service{
		repo:    repo,
		proc:    proc,
		enc:     enc,
		dataDir: dataDir,
	}
}

// SetAuditor attaches an audit recorder to the service.
func (s *Service) SetAuditor(auditor *audit.Service) {
	s.audit = auditor
}

// SetLogStreamer attaches a log streamer to the service.
func (s *Service) SetLogStreamer(streamer LogStreamer) {
	s.logStreams = streamer
}

func (s *Service) recordAudit(ctx context.Context, actorID, action, entityType, entityID string, metadata map[string]any) {
	if s.audit == nil {
		return
	}
	if err := s.audit.Record(ctx, actorID, action, entityType, entityID, metadata); err != nil {
		slog.Warn("audit write failed", "action", action, "entity_type", entityType, "entity_id", entityID, "error", err)
	}
}

// Create creates a new application record.
func (s *Service) Create(userID string, req CreateAppRequest) (*domain.App, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("app: name is required")
	}
	projectID := req.ProjectID
	if projectID == "" {
		projectID = "default"
	}

	runtime := req.Runtime
	startCmd := req.StartCmd

	app, err := s.repo.CreateForUser(userID, projectID, req.Name, req.GitURL, runtime, startCmd)
	if err != nil {
		return nil, err
	}
	s.recordAudit(context.Background(), userID, "app.create", "app", app.ID, map[string]any{
		"project_id": app.ProjectID,
		"name":       app.Name,
	})
	return app, nil
}

// Get retrieves an app by ID for the given user.
func (s *Service) Get(userID, id string) (*domain.App, error) {
	app, err := s.repo.FindByIDForUser(userID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return app, nil
}

// ListAll returns apps visible to the given user.
func (s *Service) ListAll(userID string) ([]domain.App, error) {
	return s.repo.ListAllForUser(userID)
}

// ListByProject returns all apps in a project.
func (s *Service) ListByProject(userID, projectID string) ([]domain.App, error) {
	return s.repo.ListByProjectForUser(userID, projectID)
}

// Deploy triggers a deployment for the given app.
func (s *Service) Deploy(ctx context.Context, userID, appID string) (*domain.App, error) {
	app, err := s.repo.FindByIDForUser(userID, appID)
	if err != nil {
		return nil, ErrNotFound
	}

	if err := s.repo.UpdateStatus(appID, "deploying"); err != nil {
		return nil, err
	}
	s.recordAudit(ctx, userID, "app.deploy.start", "app", appID, map[string]any{"status": "deploying"})

	deployment, err := s.repo.CreateDeployment(appID, "latest")
	if err != nil {
		s.repo.UpdateStatus(appID, "failed")
		return nil, err
	}
	_ = s.repo.UpdateDeploymentStatus(deployment.ID, "running", "Preparing deployment workspace")

	workDir, err := s.prepareAppDir(appID)
	if err != nil {
		return nil, s.failDeployment(appID, deployment.ID, fmt.Sprintf("prepare workspace: %v", err))
	}

	if s.logStreams != nil {
		s.logStreams.WatchApp(appID)
	}

	logParts := []string{fmt.Sprintf("Preparing deployment in %s", workDir)}
	if err := s.fetchSource(ctx, app.GitURL, workDir); err != nil {
		return nil, s.failDeployment(appID, deployment.ID, fmt.Sprintf("clone: %v", err))
	}
	if app.GitURL != "" {
		logParts = append(logParts, fmt.Sprintf("Cloned source from %s", app.GitURL))
	}

	runtimeResult, buildCommand, startCommand, err := s.resolveRuntimeConfig(app, workDir)
	if err != nil {
		return nil, s.failDeployment(appID, deployment.ID, fmt.Sprintf("resolve runtime: %v", err))
	}
	if runtimeResult.Name != "" {
		logParts = append(logParts, fmt.Sprintf("Resolved runtime: %s", runtimeResult.Name))
	}
	env, err := s.loadDeployedEnv(appID)
	if err != nil {
		return nil, s.failDeployment(appID, deployment.ID, fmt.Sprintf("load env: %v", err))
	}

	if buildCommand != "" {
		if output, err := s.runCommand(ctx, buildCommand, workDir, env); err != nil {
			return nil, s.failDeployment(appID, deployment.ID, fmt.Sprintf("build: %v\n%s", err, output))
		}
		logParts = append(logParts, fmt.Sprintf("Build succeeded: %s", buildCommand))
	}
	if startCommand == "" {
		return nil, s.failDeployment(appID, deployment.ID, "start: no start command configured")
	}

	slog.Info("deploy: starting app", "app", appID, "cmd", startCommand)
	if err := s.proc.Start(ctx, appID, startCommand, flattenEnv(env), workDir); err != nil {
		return nil, s.failDeployment(appID, deployment.ID, fmt.Sprintf("start: %v", err))
	}
	if err := s.verifyStarted(ctx, appID); err != nil {
		return nil, s.failDeployment(appID, deployment.ID, fmt.Sprintf("start verification: %v", err))
	}

	logParts = append(logParts, fmt.Sprintf("Started: %s", startCommand))
	if err := s.repo.UpdateDeploymentStatus(deployment.ID, "success", strings.Join(logParts, "\n")); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateStatus(appID, "running"); err != nil {
		return nil, err
	}

	updated, _ := s.repo.FindByIDForUser(userID, appID)
	return updated, nil
}

func (s *Service) prepareAppDir(appID string) (string, error) {
	appDir := filepath.Join(s.dataDir, "apps", appID)
	if err := os.RemoveAll(appDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}
	return appDir, nil
}

func (s *Service) fetchSource(ctx context.Context, gitURL, workDir string) error {
	if gitURL == "" {
		return nil
	}
	if err := os.RemoveAll(workDir); err != nil {
		return err
	}
	parent := filepath.Dir(workDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", gitURL, workDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\n%s", err, string(output))
	}
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("git clone produced an empty workdir")
	}
	return nil
}

func (s *Service) resolveRuntimeConfig(app *domain.App, workDir string) (runtimepkg.RuntimeResult, string, string, error) {
	detector := runtimepkg.New()
	result, err := detector.Detect(workDir, app.Runtime)
	if err != nil {
		return runtimepkg.RuntimeResult{}, "", "", err
	}
	buildCommand := ""
	if app.GitURL != "" {
		buildCommand = result.BuildCommand
	}
	startCommand := app.StartCmd
	if startCommand == "" {
		startCommand = result.StartCommand
	}
	return result, buildCommand, startCommand, nil
}

func (s *Service) runCommand(ctx context.Context, command, workDir string, env map[string]string) (string, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-lc", command)
	cmd.Dir = workDir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), flattenEnv(env)...)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (s *Service) loadDeployedEnv(appID string) (map[string]string, error) {
	stored, err := s.repo.ListEnvVarsWithValues(appID)
	if err != nil {
		return nil, err
	}
	if len(stored) == 0 {
		return nil, nil
	}
	env := make(map[string]string, len(stored))
	for _, item := range stored {
		decrypted, err := s.enc.Decrypt(item.Value)
		if err != nil {
			return nil, fmt.Errorf("decrypt %s: %w", item.Key, err)
		}
		env[item.Key] = string(decrypted)
	}
	return env, nil
}

func flattenEnv(env map[string]string) []string {
	pairs := make([]string, 0, len(env))
	for k, v := range env {
		pairs = append(pairs, k+"="+v)
	}
	return pairs
}

func (s *Service) verifyStarted(ctx context.Context, appID string) error {
	time.Sleep(250 * time.Millisecond)
	state, err := s.proc.Status(ctx, appID)
	if err != nil {
		return err
	}
	if state != process.StateRunning {
		return fmt.Errorf("process not running after start (state=%s)", state)
	}
	return nil
}

func (s *Service) failDeployment(appID, deploymentID, message string) error {
	_ = s.repo.UpdateDeploymentStatus(deploymentID, "failed", message)
	_ = s.repo.UpdateStatus(appID, "failed")
	return fmt.Errorf("app: deploy failed: %s", message)
}

// Stop stops a running application.
func (s *Service) Stop(ctx context.Context, userID, appID string) (*domain.App, error) {
	if _, err := s.repo.FindByIDForUser(userID, appID); err != nil {
		return nil, ErrNotFound
	}
	if err := s.proc.Stop(ctx, appID); err != nil {
		return nil, fmt.Errorf("app: stop: %w", err)
	}
	if err := s.repo.UpdateStatus(appID, "stopped"); err != nil {
		return nil, err
	}
	s.recordAudit(ctx, userID, "app.stop", "app", appID, map[string]any{"status": "stopped"})
	return s.repo.FindByIDForUser(userID, appID)
}

// Restart restarts a running application.
func (s *Service) Restart(ctx context.Context, userID, appID string) (*domain.App, error) {
	app, err := s.repo.FindByIDForUser(userID, appID)
	if err != nil {
		return nil, ErrNotFound
	}

	if err := s.proc.Restart(ctx, appID); err != nil {
		return nil, fmt.Errorf("app: restart: %w", err)
	}
	s.recordAudit(ctx, userID, "app.restart", "app", appID, map[string]any{"status": "restarting"})

	startCmd := app.StartCmd
	if startCmd == "" {
		startCmd = "echo 'no start command'"
	}

	appDir := filepath.Join(s.dataDir, "apps", appID)
	env, err := s.loadDeployedEnv(appID)
	if err != nil {
		s.repo.UpdateStatus(appID, "failed")
		return nil, fmt.Errorf("app: restart load env: %w", err)
	}
	if s.logStreams != nil {
		s.logStreams.WatchApp(appID)
	}
	if err := s.proc.Start(ctx, appID, startCmd, flattenEnv(env), appDir); err != nil {
		s.repo.UpdateStatus(appID, "failed")
		return nil, fmt.Errorf("app: restart start: %w", err)
	}

	s.repo.UpdateStatus(appID, "running")
	return s.repo.FindByIDForUser(userID, appID)
}

// Delete removes an application.
func (s *Service) Delete(ctx context.Context, userID, appID string) error {
	if _, err := s.repo.FindByIDForUser(userID, appID); err != nil {
		return ErrNotFound
	}
	_ = s.proc.Stop(ctx, appID)
	if s.logStreams != nil {
		s.logStreams.UnwatchApp(appID)
	}
	s.recordAudit(ctx, userID, "app.delete", "app", appID, nil)
	return s.repo.Delete(appID)
}

// SetEnvVars sets environment variables for an app (encrypted at rest).
func (s *Service) SetEnvVars(userID, appID string, vars []EnvVarInput) error {
	if _, err := s.repo.FindByIDForUser(userID, appID); err != nil {
		return ErrNotFound
	}
	for _, v := range vars {
		encrypted, err := s.enc.Encrypt([]byte(v.Value))
		if err != nil {
			return fmt.Errorf("app: encrypt env var: %w", err)
		}
		if err := s.repo.UpsertEnvVar(appID, v.Key, encrypted); err != nil {
			return err
		}
	}
	s.recordAudit(context.Background(), userID, "app.env.update", "app", appID, map[string]any{"keys": envKeys(vars)})
	return nil
}

// GetEnvVarKeys returns the list of env var keys (without values).
func (s *Service) GetEnvVarKeys(userID, appID string) ([]domain.AppEnvVar, error) {
	if _, err := s.repo.FindByIDForUser(userID, appID); err != nil {
		return nil, ErrNotFound
	}
	return s.repo.ListEnvVars(appID)
}

// ListDeployments returns deployments for an app.
func (s *Service) ListDeployments(userID, appID string) ([]domain.AppDeployment, error) {
	if _, err := s.repo.FindByIDForUser(userID, appID); err != nil {
		return nil, ErrNotFound
	}
	return s.repo.ListDeployments(appID)
}

func envKeys(vars []EnvVarInput) []string {
	keys := make([]string, 0, len(vars))
	for _, v := range vars {
		if v.Key != "" {
			keys = append(keys, v.Key)
		}
	}
	return keys
}
