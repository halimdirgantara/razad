package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/razad/razad/internal/crypto"
	"github.com/razad/razad/internal/domain"
	"github.com/razad/razad/internal/process"
)

var (
	ErrNotFound    = errors.New("app: not found")
	ErrNotRunning  = errors.New("app: not running")
	ErrAccessDenied = errors.New("app: access denied")
)

// Service handles app management business logic.
type Service struct {
	repo    *Repository
	proc    process.Runner
	enc     *crypto.Encrypter
	dataDir string
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

// Create creates a new application record.
func (s *Service) Create(req CreateAppRequest) (*domain.App, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("app: name is required")
	}
	if req.ProjectID == "" {
		return nil, fmt.Errorf("app: project_id is required")
	}

	runtime := req.Runtime
	startCmd := req.StartCmd

	return s.repo.Create(req.ProjectID, req.Name, req.GitURL, runtime, startCmd)
}

// Get retrieves an app by ID.
func (s *Service) Get(id string) (*domain.App, error) {
	return s.repo.FindByID(id)
}

// ListAll returns all apps across all projects.
func (s *Service) ListAll() ([]domain.App, error) {
	return s.repo.ListAll()
}

// ListByProject returns all apps in a project.
func (s *Service) ListByProject(projectID string) ([]domain.App, error) {
	return s.repo.ListByProject(projectID)
}

// Deploy triggers a deployment for the given app.
func (s *Service) Deploy(ctx context.Context, appID string) (*domain.App, error) {
	app, err := s.repo.FindByID(appID)
	if err != nil {
		return nil, ErrNotFound
	}

	// Update status to deploying
	if err := s.repo.UpdateStatus(appID, "deploying"); err != nil {
		return nil, err
	}

	// Create deployment record
	deployment, err := s.repo.CreateDeployment(appID, "latest")
	if err != nil {
		s.repo.UpdateStatus(appID, "failed")
		return nil, err
	}

	appDir := filepath.Join(s.dataDir, "apps", appID)
	workDir := appDir

	// Ensure app directory exists
	if err := os.MkdirAll(workDir, 0755); err != nil {
		s.repo.UpdateDeploymentStatus(deployment.ID, "failed", fmt.Sprintf("mkdir: %v", err))
		s.repo.UpdateStatus(appID, "failed")
		return nil, fmt.Errorf("app: create dir: %w", err)
	}

	startCommand := app.StartCmd
	if startCommand == "" {
		startCommand = "echo 'no start command configured'"
	}

	// Start the process
	logLine := fmt.Sprintf("Starting: %s in %s", startCommand, workDir)
	slog.Info("deploy: starting app", "app", appID, "cmd", startCommand)

	if err := s.proc.Start(ctx, appID, startCommand, nil, workDir); err != nil {
		s.repo.UpdateDeploymentStatus(deployment.ID, "failed", fmt.Sprintf("start: %v", err))
		s.repo.UpdateStatus(appID, "failed")
		return nil, fmt.Errorf("app: start: %w", err)
	}

	s.repo.UpdateDeploymentStatus(deployment.ID, "success", logLine)
	s.repo.UpdateStatus(appID, "running")

	updated, _ := s.repo.FindByID(appID)
	return updated, nil
}

// Stop stops a running application.
func (s *Service) Stop(ctx context.Context, appID string) (*domain.App, error) {
	if err := s.proc.Stop(ctx, appID); err != nil {
		return nil, fmt.Errorf("app: stop: %w", err)
	}

	if err := s.repo.UpdateStatus(appID, "stopped"); err != nil {
		return nil, err
	}

	return s.repo.FindByID(appID)
}

// Restart restarts a running application.
func (s *Service) Restart(ctx context.Context, appID string) (*domain.App, error) {
	app, err := s.repo.FindByID(appID)
	if err != nil {
		return nil, ErrNotFound
	}

	if err := s.proc.Restart(ctx, appID); err != nil {
		return nil, fmt.Errorf("app: restart: %w", err)
	}

	// Re-start after restart (Restart just stops in local runner)
	startCmd := app.StartCmd
	if startCmd == "" {
		startCmd = "echo 'no start command'"
	}

	appDir := filepath.Join(s.dataDir, "apps", appID)
	if err := s.proc.Start(ctx, appID, startCmd, nil, appDir); err != nil {
		s.repo.UpdateStatus(appID, "failed")
		return nil, fmt.Errorf("app: restart start: %w", err)
	}

	s.repo.UpdateStatus(appID, "running")
	return s.repo.FindByID(appID)
}

// Delete removes an application.
func (s *Service) Delete(ctx context.Context, appID string) error {
	_ = s.proc.Stop(ctx, appID)
	return s.repo.Delete(appID)
}

// SetEnvVars sets environment variables for an app (encrypted at rest).
func (s *Service) SetEnvVars(appID string, vars []EnvVarInput) error {
	for _, v := range vars {
		encrypted, err := s.enc.Encrypt([]byte(v.Value))
		if err != nil {
			return fmt.Errorf("app: encrypt env var: %w", err)
		}
		if err := s.repo.UpsertEnvVar(appID, v.Key, encrypted); err != nil {
			return err
		}
	}
	return nil
}

// GetEnvVarKeys returns the list of env var keys (without values).
func (s *Service) GetEnvVarKeys(appID string) ([]domain.AppEnvVar, error) {
	return s.repo.ListEnvVars(appID)
}
