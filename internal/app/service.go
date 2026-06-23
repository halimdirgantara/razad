package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/razad/razad/internal/audit"
	"github.com/razad/razad/internal/crypto"
	"github.com/razad/razad/internal/domain"
	"github.com/razad/razad/internal/process"
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

	appDir := filepath.Join(s.dataDir, "apps", appID)
	workDir := appDir

	if s.logStreams != nil {
		s.logStreams.WatchApp(appID)
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		s.repo.UpdateDeploymentStatus(deployment.ID, "failed", fmt.Sprintf("mkdir: %v", err))
		s.repo.UpdateStatus(appID, "failed")
		return nil, fmt.Errorf("app: create dir: %w", err)
	}

	startCommand := app.StartCmd
	if startCommand == "" {
		startCommand = "echo 'no start command configured'"
	}

	logLine := fmt.Sprintf("Starting: %s in %s", startCommand, workDir)
	slog.Info("deploy: starting app", "app", appID, "cmd", startCommand)

	if err := s.proc.Start(ctx, appID, startCommand, nil, workDir); err != nil {
		s.repo.UpdateDeploymentStatus(deployment.ID, "failed", fmt.Sprintf("start: %v", err))
		s.repo.UpdateStatus(appID, "failed")
		return nil, fmt.Errorf("app: start: %w", err)
	}

	s.repo.UpdateDeploymentStatus(deployment.ID, "success", logLine)
	s.repo.UpdateStatus(appID, "running")

	updated, _ := s.repo.FindByIDForUser(userID, appID)
	return updated, nil
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
	if s.logStreams != nil {
		s.logStreams.WatchApp(appID)
	}
	if err := s.proc.Start(ctx, appID, startCmd, nil, appDir); err != nil {
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
